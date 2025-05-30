package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	httpHelpers "github.com/jacobmcgowan/simple-scheduler/services/cli/http-helpers"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
)

type JobService struct {
	ApiUrl      string
	AccessToken string
}

func (svc JobService) Browse() ([]dtos.Job, error) {
	url := fmt.Sprintf("%s/jobs", svc.ApiUrl)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", svc.AccessToken))
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, httpHelpers.ParseError(resp, "failed to get jobs")
	}

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

func (svc JobService) Read(name string) (dtos.Job, error) {
	url := fmt.Sprintf("%s/jobs/%s", svc.ApiUrl, name)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return dtos.Job{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", svc.AccessToken))
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return dtos.Job{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return dtos.Job{}, httpHelpers.ParseError(resp, "failed to get job")
	}

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

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", svc.AccessToken))
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", httpHelpers.ParseError(resp, "failed to add job")
	}

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

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", svc.AccessToken))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return httpHelpers.ParseError(resp, "failed to update job")
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
