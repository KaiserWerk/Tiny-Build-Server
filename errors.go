package main

type buildDefinitionNotFound struct {
	Id string
}

func (e buildDefinitionNotFound) Error() string {
	return "build definition directory with Id " + e.Id + " not found"
}

type buildDefinitionConfigFileNotFound struct {
	Id string
}

func (e buildDefinitionConfigFileNotFound) Error() string {
	return "config file of build definition with Id " + e.Id + " not found"
}
