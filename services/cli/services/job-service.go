package services

import (
	"bytes"
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

func (svc JobService) Add(job dtos.Job) (string, error) {
	url := fmt.Sprintf("%s/jobs", svc.ApiUrl)
	reqBody, err := json.Marshal(job)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var addedJob dtos.Job
	err = json.Unmarshal(respBody, &addedJob)
	if err != nil {
		return "", err
	}

	return addedJob.Name, nil
}

func (svc JobService) Edit(name string, jobUpdate dtos.JobUpdate) error {
	url := fmt.Sprintf("%s/jobs/%s", svc.ApiUrl, name)
	reqBody, err := json.Marshal(jobUpdate)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
