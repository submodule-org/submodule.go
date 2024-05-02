package submodule_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/submodule-org/submodule.go"
)

func TestCreate(t *testing.T) {
	var config = submodule.Create(func(ctx context.Context) (string, error) {
		return "myconfig", nil
	})

	result, err := config.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, "myconfig", result)
}

func TestSubmodule(t *testing.T) {

	t.Run("Can call object in singleton mode", func(t *testing.T) {
		ctx := context.Background()
		defer submodule.DisposeLegacyStore(ctx)

		count := 0
		counter := submodule.Create(func(ctx context.Context) (x any, e error) {
			count = count + 1
			return
		})

		counter.Get(ctx)
		counter.Get(ctx)
		counter.Get(ctx)

		require.Equal(t, count, 1)
	})

	t.Run("submodule.Derive can also be singleton", func(t *testing.T) {
		count := 0
		counter := submodule.Create(func(ctx context.Context) (x int, e error) {
			count = count + 1
			return count, nil
		})

		derivedCount := 0
		derived := submodule.Derive(func(ctx context.Context, count int) (x int, e error) {
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
		var config = submodule.Create(func(ctx context.Context) (string, error) {
			return "myconfig", nil
		})
		result, err := submodule.Execute(context.Background(), func(ctx context.Context, v string) (string, error) {
			return v + "executed", nil
		}, config)
		require.NoError(t, err)
		require.Equal(t, "myconfigexecuted", result)
	})

	t.Run("derive2 should work", func(t *testing.T) {
		var config1 = submodule.Create(func(ctx context.Context) (string, error) {
			return "myconfig1", nil
		})
		var config2 = submodule.Create(func(ctx context.Context) (int, error) {
			return 2, nil
		})
		sub := submodule.Derive2(func(ctx context.Context, dep1 string, dep2 int) (string, error) {
			return dep1 + fmt.Sprintf("%d", dep2), nil
		}, config1, config2)
		result, err := sub.Get(context.Background())
		require.NoError(t, err)
		require.Equal(t, "myconfig12", result)
	})

	t.Run("flow should work", func(t *testing.T) {
		type Config struct {
			port string
		}

		config := submodule.Create(func(ctx context.Context) (Config, error) {
			return Config{
				port: "4000",
			}, nil
		})

		type Handler = func(ctx context.Context, runtime string) string

		concatConfig := submodule.Derive(func(ctx context.Context, config Config) (Handler, error) {
			return func(ctx context.Context, r string) string {
				return config.port + r
			}, nil
		}, config)

		handler, _ := concatConfig.Get(context.Background())
		r := handler(context.Background(), "something")

		require.Equal(t, r, "4000something")
	})

	t.Run("should handle panic", func(t *testing.T) {
		withNormal := submodule.Create(func(ctx context.Context) (x string, e error) {
			return "normal", nil
		})

		withError := submodule.Create(func(ctx context.Context) (x string, e error) {
			return x, errors.New("1")
		})

		derivedFromNormal := submodule.Derive(func(ctx context.Context, x string) (string, error) {
			return x, errors.New("2")
		}, withNormal)

		derivedFromError := submodule.Derive(func(ctx context.Context, x string) (string, error) {
			return "", errors.New("3")
		}, withError)

		_, e := derivedFromNormal.Get(context.Background())

		require.NotNil(t, e)
		require.Equal(t, e.Error(), "2")

		_, e = derivedFromError.Get(context.Background())
		require.NotNil(t, e)
		require.Equal(t, e.Error(), "1")

	})

	t.Run("can replace value using context", func(t *testing.T) {
		intValue := submodule.Make[int](func() int { return 100 })
		ctx := context.Background()

		ctx = context.WithValue(ctx, intValue, 300)
		derivedIntValue := submodule.Make[int](func(v int) int {
			return v + 100
		}, intValue)

		v, e := derivedIntValue.Get(ctx)
		require.Nil(t, e)
		require.Equal(t, 400, v)
	})
}

func TestDeriveSingleton(t *testing.T) {
	config := submodule.Create(func(ctx context.Context) (string, error) {
		return "myconfig", nil
	})
	derived := submodule.Derive(func(ctx context.Context, d string) (string, error) {
		return d + "derived", nil
	}, config)
	result, err := derived.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, "myconfigderived", result)

	derived1 := submodule.Derive(func(ctx context.Context, d string) (*string, error) {
		v := d + "derived"
		return &v, nil
	}, config)
	result1, err := derived1.Get(context.Background())
	require.NoError(t, err)
	result2, err := derived1.Get(context.Background())
	require.NoError(t, err)

	require.True(t, result1 == result2) // pointer comparison
}
