package common

type EntryType string

const (
	DocumentType    EntryType = "DocumentType"
	CollectionType  EntryType = "CollectionType"
	TemplateType    EntryType = "TemplateType" // reMarkable Methods (e.g. source com.remarkable.methods)
)
