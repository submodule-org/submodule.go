package submodule

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	var config = Create(func(ctx context.Context) (string, error) {
		return "myconfig", nil
	})

	result, err := config.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, "myconfig", result)
}

func TestSubmodule(t *testing.T) {

	t.Run("Can call object in singleton mode", func(t *testing.T) {
		count := 0
		counter := Create(func(ctx context.Context) (x any, e error) {
			count = count + 1
			return
		})

		counter.Get(context.TODO())
		counter.Get(context.TODO())
		counter.Get(context.TODO())

		require.Equal(t, count, 1)
	})

	t.Run("Can call object in prototype mode", func(t *testing.T) {
		count := 0
		counter := Create(func(ctx context.Context) (x any, e error) {
			count = count + 1
			return
		}, SetPrototype)

		counter.Get(context.TODO())
		counter.Get(context.TODO())
		counter.Get(context.TODO())

		require.Equal(t, count, 3)
	})

	t.Run("Derive would still hornor running mode", func(t *testing.T) {
		count := 0
		counter := Create(func(ctx context.Context) (x int, e error) {
			count = count + 1
			return count, nil
		})

		derivedCount := 0
		derived := Derive(func(ctx context.Context, count int) (x int, e error) {
			derivedCount = derivedCount + 1
			return count, nil
		}, counter, SetPrototype)

		derived.Get(context.TODO())
		derived.Get(context.TODO())
		derived.Get(context.TODO())

		require.Equal(t, derivedCount, 3)
		require.Equal(t, count, 3)
	})

	t.Run("Derive can also be singleton", func(t *testing.T) {
		count := 0
		counter := Create(func(ctx context.Context) (x int, e error) {
			count = count + 1
			return count, nil
		})

		derivedCount := 0
		derived := Derive(func(ctx context.Context, count int) (x int, e error) {
			derivedCount = derivedCount + 1
			return count, nil
		}, counter)

		derived.Get(context.TODO())
		derived.Get(context.TODO())
		derived.Get(context.TODO())

		require.Equal(t, derivedCount, 1)
		require.Equal(t, count, 1)
	})

	t.Run("execute should work", func(t *testing.T) {
		var config = Create(func(ctx context.Context) (string, error) {
			return "myconfig", nil
		})
		result, err := Execute(context.Background(), func(ctx context.Context, v string) (string, error) {
			return v + "executed", nil
		}, config)
		require.NoError(t, err)
		require.Equal(t, "myconfigexecuted", result)
	})

	t.Run("derive2 should work", func(t *testing.T) {
		var config1 = Create(func(ctx context.Context) (string, error) {
			return "myconfig1", nil
		})
		var config2 = Create(func(ctx context.Context) (string, error) {
			return "myconfig2", nil
		})
		sub := Derive2(func(ctx context.Context, dep1, dep2 string) (string, error) {
			return dep1 + dep2, nil
		}, config1, config2)
		result, err := sub.Get(context.Background())
		require.NoError(t, err)
		require.Equal(t, "myconfig1myconfig2", result)
	})

	t.Run("flow should work", func(t *testing.T) {
		type Config struct {
			port string
		}

		config := Create(func(ctx context.Context) (Config, error) {
			return Config{
				port: "4000",
			}, nil
		})

		type Handler = func(ctx context.Context, runtime string) string

		concatConfig := Derive(func(ctx context.Context, config Config) (Handler, error) {
			return func(ctx context.Context, r string) string {
				return config.port + r
			}, nil
		}, config)

		handler, _ := concatConfig.Get(context.Background())
		r := handler(context.Background(), "something")

		require.Equal(t, r, "4000something")
	})

	t.Run("should be able to replace value to test", func(t *testing.T) {
		ctx := context.Background()

		config := Create(func(ctx context.Context) (string, error) {
			return "myconfig1", nil
		})

		configMock := Create(func(ctx context.Context) (string, error) {
			return "myconfig_mock_", nil
		})

		derived := Derive(func(ctx context.Context, d string) (string, error) {
			return d + "derived", nil
		}, config)
		ctx = context.WithValue(ctx, config, configMock)
		result, err := derived.Get(ctx)
		require.NoError(t, err)
		require.Equal(t, "myconfig_mock_derived", result)
	})

	t.Run("should handle panic", func(t *testing.T) {
		withNormal := Create(func(ctx context.Context) (x string, e error) {
			return "normal", nil
		})

		withError := Create(func(ctx context.Context) (x string, e error) {
			return x, errors.New("1")
		})

		derivedFromNormal := Derive(func(ctx context.Context, x string) (string, error) {
			return x, errors.New("2")
		}, withNormal)

		derivedFromError := Derive(func(ctx context.Context, x string) (string, error) {
			return "", errors.New("3")
		}, withError)

		_, e := derivedFromNormal.Get(context.Background())

		require.NotNil(t, e)
		require.Equal(t, e.Error(), "2")

		_, e = derivedFromError.Get(context.Background())
		require.NotNil(t, e)
		require.Equal(t, e.Error(), "1")

	})

	t.Run("what if safe flow causes panic?", func(t *testing.T) {

	})

}

func TestDerive3(t *testing.T) {
	var config1 = Create(func(ctx context.Context) (string, error) {
		return "myconfig1", nil
	})
	var config2 = Create(func(ctx context.Context) (string, error) {
		return "myconfig2", nil
	})
	var config3 = Create(func(ctx context.Context) (string, error) {
		return "myconfig3", nil
	})
	sub := Derive3(func(ctx context.Context, dep1, dep2, dep3 string) (string, error) {
		return dep1 + dep2 + dep3, nil
	}, config1, config2, config3)
	result, err := sub.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, "myconfig1myconfig2myconfig3", result)
}
func Test_Support_Cancel_Context(t *testing.T) {
	var config = Create(func(ctx context.Context) (string, error) {
		return "myconfig", nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := config.Get(ctx)
	require.Error(t, err)
	require.Equal(t, context.Canceled, err)
}

func TestDeriveSingleton(t *testing.T) {
	config := Create(func(ctx context.Context) (string, error) {
		return "myconfig", nil
	})
	derived := Derive(func(ctx context.Context, d string) (string, error) {
		return d + "derived", nil
	}, config)
	result, err := derived.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, "myconfigderived", result)

	derived1 := Derive(func(ctx context.Context, d string) (*string, error) {
		v := d + "derived"
		return &v, nil
	}, config)
	result1, err := derived1.Get(context.Background())
	require.NoError(t, err)
	result2, err := derived1.Get(context.Background())
	require.NoError(t, err)

	require.True(t, result1 == result2) // pointer comparison
}

func TestDerive_Prototype(t *testing.T) {
	t.Setenv("ENV", "dev")
	config := Create(func(ctx context.Context) (string, error) {
		return "myconfig" + os.Getenv("ENV"), nil
	})
	derived := Derive(func(ctx context.Context, d string) (string, error) {
		return d + "derived", nil
	}, config, SetPrototype)

	result1, err := derived.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, "myconfigdevderived", result1)

	t.Setenv("ENV", "prod")
	result2, err := derived.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, "myconfigprodderived", result2)
}

func TestPrestage(t *testing.T) {
	type Config struct {
		port string
	}

	type Service struct {
		service string
	}

	config := Create(func(ctx context.Context) (Config, error) {
		return Config{
			port: "4000",
		}, nil
	})

	mod := Prestage(func(ctx context.Context, config Config) (Service, error) {
		return Service{
			service: config.port,
		}, nil
	})

	service := mod(config)
	port, _ := service.Get(context.TODO())
	require.Equal(t, port.service, "4000")
}
