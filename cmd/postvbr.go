package cmd

import (
	"io/ioutil"
	"os"

	"github.com/shapedthought/veeamcli/models"
	"github.com/shapedthought/veeamcli/utils"
	"github.com/shapedthought/veeamcli/vbrmodels"
	"github.com/shapedthought/veeamcli/vhttp"
	"gopkg.in/yaml.v2"
)

func vbrCreateJob(profile models.Profile, path string) {

	var jobdata vbrmodels.VbrJob

	yml, err := os.Open(path)
	utils.IsErr(err)

	b, err := ioutil.ReadAll(yml)
	utils.IsErr(err)

	err = yaml.Unmarshal(b, &jobdata)
	utils.IsErr(err)

	jc := vhttp.PostData("jobs", profile, jobdata)

	if jc {
		print("job created")
	}

}
