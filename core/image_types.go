package core

type ImageVersionInfo struct {
	Name               string `json:"name" jsonschema:"the version name/tag"`
	PythonDependencies string `json:"pythonDependencies,omitempty" jsonschema:"Python dependencies if available"`
	Software           string `json:"software,omitempty" jsonschema:"Software information if available"`
}

type ImageInfo struct {
	Name     string             `json:"name" jsonschema:"the display name of the image"`
	URL      string             `json:"url" jsonschema:"the URL/location of the image"`
	Versions []ImageVersionInfo `json:"versions" jsonschema:"available versions of the image"`
}

type ListImagesOutput struct {
	Images []ImageInfo `json:"images" jsonschema:"the list of images"`
}

type CreateCustomImageInput struct {
	ImageLocation    string `json:"imageLocation" jsonschema:"the location of the image"`
	ImageName        string `json:"imageName" jsonschema:"the name of the image"`
	ImageDescription string `json:"imageDescription" jsonschema:"the description of the image"`
}

type UpdateImageInput struct {
	ImageName        string `json:"imageName" jsonschema:"the name of the image to update"`
	NewImageName     string `json:"newImageName,omitempty" jsonschema:"optional - new name to rename the image to"`
	ImageDescription string `json:"imageDescription,omitempty" jsonschema:"optional - new description for the image"`
}

type DeleteImageInput struct {
	ImageName string `json:"imageName" jsonschema:"the name of the image"`
}
