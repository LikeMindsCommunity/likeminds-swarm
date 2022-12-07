package entities

type ImageWidget struct {
	FileType int    `json:"file_type" bson:"file_type"`
	FileUrl  string `json:"file_url" bson:"file_url"`
}

type VideoWidget struct {
	FileType int    `json:"file_type" bson:"file_type"`
	FileUrl  string `json:"file_url" bson:"file_url"`
}

type DocumentWidget struct {
	FileType   int    `json:"file_type" bson:"file_type"`
	FileUrl    string `json:"file_url" bson:"file_url"`
	FileFormat string `json:"file_format" bson:"file_format"`
	FileSize   string `json:"file_size" bson:"file_size"`
}

func NewImageWidget(file_type int, file_url string) ImageWidget {
	return ImageWidget{
		FileType: file_type,
		FileUrl:  file_url,
	}
}

func NewVideoWidget(file_type int, file_url string) VideoWidget {
	return VideoWidget{
		FileType: file_type,
		FileUrl:  file_url,
	}
}

func NewDocumentWidget(file_type int, file_url string, file_format string, file_size string) DocumentWidget {
	return DocumentWidget{
		FileType:   file_type,
		FileUrl:    file_url,
		FileFormat: file_format,
		FileSize:   file_size,
	}
}
