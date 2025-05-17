package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	httpHelpers "github.com/jacobmcgowan/simple-scheduler/services/cli/http-helpers"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
)

type RunService struct {
	ApiUrl      string
	AccessToken string
}

func (svc RunService) Browse(filter dtos.RunFilter) ([]dtos.Run, error) {
	qb := httpHelpers.NewQueryBuilder()
	qb.Add("jobName", filter.JobName)
	qb.Add("status", (*string)(filter.Status))

	url := fmt.Sprintf("%s/runs%s", svc.ApiUrl, qb.String())
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
		return nil, httpHelpers.ParseError(resp, "failed to get runs")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var runs []dtos.Run
	err = json.Unmarshal(body, &runs)
	if err != nil {
		return nil, err
	}

	return runs, nil
}

func (svc RunService) Read(id string) (dtos.Run, error) {
	url := fmt.Sprintf("%s/runs/%s", svc.ApiUrl, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return dtos.Run{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", svc.AccessToken))
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return dtos.Run{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return dtos.Run{}, httpHelpers.ParseError(resp, "failed to get run")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return dtos.Run{}, err
	}

	var run dtos.Run
	err = json.Unmarshal(body, &run)
	if err != nil {
		return dtos.Run{}, err
	}

	return run, nil
}

func (svc RunService) Cancel(id string) error {
	url := fmt.Sprintf("%s/runs/%s", svc.ApiUrl, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", svc.AccessToken))
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return httpHelpers.ParseError(resp, "failed to cancel run")
	}

	return nil
}
