package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	httpHelpers "github.com/jacobmcgowan/simple-scheduler/services/cli/http-helpers"
	"github.com/jacobmcgowan/simple-scheduler/shared/common"
	"github.com/jacobmcgowan/simple-scheduler/shared/dtos"
)

type RunService struct {
	ApiUrl string
}

func (srv RunService) Browse(filter dtos.RunFilter) ([]dtos.Run, error) {
	status := common.Undefinable[string]{
		Value:   (string)(filter.Status.Value),
		Defined: filter.Status.Defined,
	}
	qb := httpHelpers.NewQueryBuilder()
	qb.Add("jobName", filter.JobName)
	qb.Add("status", status)

	url := fmt.Sprintf("%s/runs%s", srv.ApiUrl, qb.String())
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
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

func (srv RunService) Read(id string) (dtos.Run, error) {
	url := fmt.Sprintf("%s/runs/%s", srv.ApiUrl, id)
	resp, err := http.Get(url)
	if err != nil {
		return dtos.Run{}, err
	}

	defer resp.Body.Close()
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

func (srv RunService) Cancel(id string) error {
	url := fmt.Sprintf("%s/runs/%s", srv.ApiUrl, id)
	_, err := http.Get(url)
	if err != nil {
		return err
	}

	return nil
}
