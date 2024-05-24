module github.com/submodule-org/submodule.go/batteries/sub_cmd

go 1.22.2

replace (
	github.com/submodule-org/submodule.go => ../..
	github.com/submodule-org/submodule.go/batteries/sub_env => ../sub_env
)

require (
	github.com/spf13/cobra v1.8.0
	github.com/submodule-org/submodule.go v1.7.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)
