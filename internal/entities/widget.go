package entities

type Widget struct {
	FileType   int    `json:"file_type" bson:"file_type"`
	FileUrl    string `json:"file_url,omitempty" bson:"file_url,omitempty"`
	FileFormat string `json:"file_format,omitempty" bson:"file_format,omitempty"`
	FileSize   string `json:"file_size,omitempty" bson:"file_size,omitempty"`
}

func NewWidget(file_type int, file_url string, file_format string, file_size string) Widget {
	return Widget{
		FileType:   file_type,
		FileUrl:    file_url,
		FileFormat: file_format,
		FileSize:   file_size,
	}
}
