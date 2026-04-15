package mpr

type MxMetadata struct {
	ProductVersion string     `yaml:"ProductVersion"`
	BuildVersion   string     `yaml:"BuildVersion"`
	Modules        []MxModule `yaml:"Modules"`
}

type MxUnit struct {
	UnitID          string                 `yaml:"UnitID"`
	ContainerID     string                 `yaml:"ContainerID"`
	ContainmentName string                 `yaml:"ContainmentName"`
	Contents        map[string]interface{} `yaml:"Contents"`
	ContentsHash    string                 `yaml:"ContentsHash,omitempty"`
}

type MxDocument struct {
	Name         string                 `yaml:"Name"`
	Type         string                 `yaml:"Type"`
	Path         string                 `yaml:"Path"`
	Attributes   map[string]interface{} `yaml:"Attributes"`
	ContentsHash string                 `yaml:"ContentsHash,omitempty"`
}

type MxModule struct {
	Name                string `yaml:"Name"`
	ID                  string `yaml:"ID"`
	FromAppStore        bool   `yaml:"FromAppStore,omitempty"`
	AppStoreVersion     string `yaml:"AppStoreVersion,omitempty"`
	AppStoreGuid        string `yaml:"AppStoreGuid,omitempty"`
	AppStoreVersionGuid string `yaml:"AppStoreVersionGuid,omitempty"`
	AppStorePackageId   string `yaml:"AppStorePackageId,omitempty"`
}

type MxFolder struct {
	Name     string    `yaml:"Name"`
	ID       string    `yaml:"ID"`
	ParentID string    `yaml:"ParentID"`
	Parent   *MxFolder `yaml:"Parent"`
}

type MxID struct {
	Data    string `yaml:"Data"`
	Subtype int    `yaml:"Subtype"`
}
