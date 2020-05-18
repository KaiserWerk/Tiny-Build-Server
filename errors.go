package main

type BuildDefinitionNotFound struct{
	Id	string
}

func (e BuildDefinitionNotFound) Error() string {
	return "build definition directory with id " + e.Id + " not found"
}

type BuildDefinitionConfigFileNotFound struct{
	Id	string
}

func (e BuildDefinitionConfigFileNotFound) Error() string {
	return "config file of build definition with id " + e.Id + " not found"
}
