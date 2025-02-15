// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package kibana

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/elastic/elastic-package/internal/logger"
)

const findDashboardsPerPage = 100

// DashboardSavedObject corresponds to the Kibana dashboard saved object
type DashboardSavedObject struct {
	ID    string
	Title string
}

// DashboardSavedObjects is an array of DashboardSavedObject
type DashboardSavedObjects []DashboardSavedObject

type savedObjectsResponse struct {
	Total        int
	SavedObjects []savedObjectResponse `json:"saved_objects"`

	Error   string
	Message string
}

type savedObjectResponse struct {
	ID         string
	Attributes struct {
		Title string
	}
}

// Strings method returns string representation for a set of saved objects.
func (dsos DashboardSavedObjects) Strings() []string {
	var entries []string
	for _, dso := range dsos {
		entries = append(entries, dso.String())
	}
	return entries
}

// String method returns a string representation for Kibana dashboard saved object.
func (dso *DashboardSavedObject) String() string {
	return fmt.Sprintf("%s (ID: %s)", dso.Title, dso.ID)
}

// FindDashboards method returns dashboards available in the Kibana instance.
func (c *Client) FindDashboards() (DashboardSavedObjects, error) {
	logger.Debug("Find dashboards using the Saved Objects API")

	var foundObjects DashboardSavedObjects
	page := 1

	for {
		r, err := c.findDashboardsNextPage(page)
		if err != nil {
			return nil, errors.Wrap(err, "can't fetch page with results")
		}
		if r.Error != "" {
			return nil, fmt.Errorf("%s: %s", r.Error, r.Message)
		}

		for _, savedObject := range r.SavedObjects {
			foundObjects = append(foundObjects, DashboardSavedObject{
				ID:    savedObject.ID,
				Title: savedObject.Attributes.Title,
			})
		}

		if r.Total <= len(foundObjects) {
			break
		}
		page++
	}

	sort.Slice(foundObjects, func(i, j int) bool {
		return sort.StringsAreSorted([]string{strings.ToLower(foundObjects[i].Title), strings.ToLower(foundObjects[j].Title)})
	})
	return foundObjects, nil
}

func (c *Client) findDashboardsNextPage(page int) (*savedObjectsResponse, error) {
	path := fmt.Sprintf("%s/_find?type=dashboard&fields=title&per_page=%d&page=%d", SavedObjectsAPI, findDashboardsPerPage, page)
	statusCode, respBody, err := c.get(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not find dashboards; API status code = %d; response body = %s", statusCode, respBody)
	}

	var r savedObjectsResponse
	err = json.Unmarshal(respBody, &r)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling response failed")
	}
	return &r, nil
}
