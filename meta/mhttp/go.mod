module github.com/submodule-org/submodule.go/meta/mhttp

go 1.22.2

replace (
	github.com/submodule-org/submodule.go => ../..
	github.com/submodule-org/submodule.go/meta/menv => ../menv
)

require github.com/submodule-org/submodule.go v1.7.0
