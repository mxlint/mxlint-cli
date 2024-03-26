package mpr

type MxMetadata struct {
	ProductVersion string `yaml:"ProductVersion"`
	BuildVersion   string `yaml:"BuildVersion"`
}

type MxUnit struct {
	UnitID          string                 `yaml:"UnitID"`
	ContainerID     string                 `yaml:"ContainerID"`
	ContainmentName string                 `yaml:"ContainmentName"`
	Contents        map[string]interface{} `yaml:"Contents"`
}

type MxDocument struct {
	Name       string                 `yaml:"Name"`
	Type       string                 `yaml:"Type"`
	Path       string                 `yaml:"Path"`
	Attributes map[string]interface{} `yaml:"Attributes"`
}

type MxModule struct {
	Name       string                 `yaml:"Name"`
	ID         string                 `yaml:"ID"`
	Attributes map[string]interface{} `yaml:"Attributes"`
}

type MxFolder struct {
	Name       string                 `yaml:"Name"`
	ID         string                 `yaml:"ID"`
	ParentID   string                 `yaml:"ParentID"`
	Parent     *MxFolder              `yaml:"Parent"`
	Attributes map[string]interface{} `yaml:"Attributes"`
}
