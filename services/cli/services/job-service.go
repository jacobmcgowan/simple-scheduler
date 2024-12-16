package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
)

type JobService struct {
	ApiUrl string
}

func (srv JobService) Browse() ([]dtos.Job, error) {
	url := fmt.Sprintf("%s/jobs", srv.ApiUrl)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jobs []dtos.Job
	err = json.Unmarshal(body, &jobs)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func (srv JobService) Read(name string) (dtos.Job, error) {
	url := fmt.Sprintf("%s/jobs/%s", srv.ApiUrl, name)
	resp, err := http.Get(url)
	if err != nil {
		return dtos.Job{}, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return dtos.Job{}, err
	}

	var job dtos.Job
	err = json.Unmarshal(body, &job)
	if err != nil {
		return dtos.Job{}, err
	}

	return job, nil
}
