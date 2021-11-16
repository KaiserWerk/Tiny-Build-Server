package assets

import "testing"

func TestGetConfig(t *testing.T) {
	_, err := GetConfig("app.dist.yaml")
	if err != nil {
		t.Fatalf("could not read config file: %s", err.Error())
	}
}

func TestGetTemplate(t *testing.T) {
	_, err := GetTemplate("_layout.html")
	if err != nil {
		t.Fatalf("could not read template file: %s", err.Error())
	}
}

func TestGetWebAssetFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"404 image", "assets/error-404-monochrome.svg", false},
		{"stylesheet", "css/styles.css", false},
		{"jQuery", "js/jquery-3.5.1.min.js", false},
		{"fontawesome", "js/fontawesome-5.13.0.min.js", false},
		{"javascripts", "js/scripts.js", false},
		{"err 1", "js/script.js", true},
		{"err 2", "css/somefile.css", true},
		{"err 3", "assets/image1.png", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := GetWebAssetFile(test.filename); (err != nil) != test.wantErr {
				t.Fatalf("could not get web asset file: %s", err.Error())
			}
		})
	}
}

func TestGetMiscFile(t *testing.T) {
	_, err := GetMiscFile("build_definition_skeleton.yaml")
	if err != nil {
		t.Fatalf("could not read misc file: %s", err.Error())
	}
}
