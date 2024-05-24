module github.com/submodule-org/submodule.go/batteries/sub_http

go 1.22.2

replace (
	github.com/submodule-org/submodule.go => ../..
	github.com/submodule-org/submodule.go/batteries/sub_env => ../sub_env
)

require (
	github.com/submodule-org/submodule.go v1.7.0
	github.com/submodule-org/submodule.go/batteries/enved v0.0.0-00010101000000-000000000000
)
