package app

import (
	"reflect"
	"testing"

	appb "github.com/luizalabs/teresa/pkg/protobuf/app"
)

func newCreateRequest(processType, protocol string) *appb.CreateRequest {
	lrq1 := &appb.CreateRequest_Limits_LimitRangeQuantity{
		Quantity: "1",
		Resource: "resource1",
	}
	lrq2 := &appb.CreateRequest_Limits_LimitRangeQuantity{
		Quantity: "2",
		Resource: "resource2",
	}
	lrq3 := &appb.CreateRequest_Limits_LimitRangeQuantity{
		Quantity: "3",
		Resource: "resource3",
	}
	lrq4 := &appb.CreateRequest_Limits_LimitRangeQuantity{
		Quantity: "4",
		Resource: "resource4",
	}
	lim := &appb.CreateRequest_Limits{
		Default: []*appb.CreateRequest_Limits_LimitRangeQuantity{
			lrq1,
			lrq2,
		},
		DefaultRequest: []*appb.CreateRequest_Limits_LimitRangeQuantity{
			lrq3,
			lrq4,
		},
	}
	as := &appb.CreateRequest_Autoscale{
		CpuTargetUtilization: 42,
		Max:                  666,
		Min:                  1,
	}
	return &appb.CreateRequest{
		Name:        "name",
		Team:        "team",
		ProcessType: processType,
		VirtualHost: "test.teresa-apps.io",
		Autoscale:   as,
		Limits:      lim,
		Protocol:    protocol,
	}
}

func TestNewApp(t *testing.T) {
	req := newCreateRequest("test", "test")
	app := newApp(req)
	want := &App{
		Name:        "name",
		Team:        "team",
		ProcessType: "test",
		VirtualHost: "test.teresa-apps.io",
		Protocol:    "test",
		Autoscale: &Autoscale{
			CPUTargetUtilization: 42,
			Max:                  666,
			Min:                  1,
		},
		Limits: &Limits{
			Default: []*LimitRangeQuantity{
				{Resource: "resource1", Quantity: "1"},
				{Resource: "resource2", Quantity: "2"},
			},
			DefaultRequest: []*LimitRangeQuantity{
				{Resource: "resource3", Quantity: "3"},
				{Resource: "resource4", Quantity: "4"},
			},
		},
		EnvVars: []*EnvVar{},
	}

	if !reflect.DeepEqual(app, want) {
		t.Errorf("got %v; want %v", app, want)
	}
}

func TestNewAppDefaults(t *testing.T) {
	var testCases = []struct {
		req         *appb.CreateRequest
		processType string
		protocol    string
	}{
		{
			newCreateRequest("", ""),
			ProcessTypeWeb,
			defaultAppProtocol,
		},
		{
			newCreateRequest("", "test"),
			ProcessTypeWeb,
			"test",
		},
		{
			newCreateRequest("test", ""),
			"test",
			"",
		},
	}

	for _, tc := range testCases {
		app := newApp(tc.req)
		if app.ProcessType != tc.processType {
			t.Errorf("got %s; want %s", app.ProcessType, tc.processType)
		}
		if app.Protocol != tc.protocol {
			t.Errorf("got %s; want %s", app.Protocol, tc.protocol)
		}
	}
}

func TestNewInfoResponse(t *testing.T) {
	lrq1 := &LimitRangeQuantity{Quantity: "1", Resource: "resource1"}
	lrq2 := &LimitRangeQuantity{Quantity: "2", Resource: "resource2"}
	info := &Info{
		Team:      "luizalabs",
		Addresses: []*Address{{Hostname: "host1"}},
		EnvVars: []*EnvVar{
			{Key: "key1", Value: "value1"},
			{Key: "key2", Value: "value2"},
		},
		Status: &Status{
			CPU:  42,
			Pods: []*Pod{{Name: "pod 1", State: "Running", Age: 1000, Restarts: 42, Ready: true}},
		},
		Autoscale: &Autoscale{CPUTargetUtilization: 33, Max: 10, Min: 1},
		Limits: &Limits{
			Default:        []*LimitRangeQuantity{lrq1},
			DefaultRequest: []*LimitRangeQuantity{lrq2},
		},
		Volumes: []string{"/teresa/secret/foo.txt"},
	}
	want := &appb.InfoResponse{
		Team:      info.Team,
		Addresses: []*appb.InfoResponse_Address{{Hostname: "host1"}},
		EnvVars: []*appb.InfoResponse_EnvVar{
			{Key: "key1", Value: "value1"},
			{Key: "key2", Value: "value2"},
		},
		Status: &appb.InfoResponse_Status{
			Cpu: info.Status.CPU,
			Pods: []*appb.InfoResponse_Status_Pod{
				{Name: "pod 1", State: "Running", Age: 1000, Restarts: 42, Ready: true},
			},
		},
		Autoscale: &appb.InfoResponse_Autoscale{
			CpuTargetUtilization: info.Autoscale.CPUTargetUtilization,
			Max:                  info.Autoscale.Max,
			Min:                  info.Autoscale.Min,
		},
		Limits: &appb.InfoResponse_Limits{
			Default: []*appb.InfoResponse_Limits_LimitRangeQuantity{
				{Quantity: "1", Resource: "resource1"},
			},
			DefaultRequest: []*appb.InfoResponse_Limits_LimitRangeQuantity{
				{Quantity: "2", Resource: "resource2"},
			},
		},
		Volumes: []string{"/teresa/secret/foo.txt"},
	}

	resp := newInfoResponse(info)
	if !reflect.DeepEqual(resp, want) {
		t.Errorf("got %v; want %v", resp, want)
	}
}

func TestNewListResponse(t *testing.T) {
	items := []*AppListItem{{
		Team:      "luizalabs",
		Addresses: []*Address{{Hostname: "host1"}},
		Name:      "teste",
	}}

	resp := newListResponse(items)
	if len(items) != len(resp.Apps) {
		t.Fatalf("expected %d items, got %d", len(items), len(resp.Apps))
	}
	itemExpected := items[0]
	itemActual := resp.Apps[0]
	expectedUrl := itemExpected.Addresses[0].Hostname
	actualUrl := itemActual.Urls[0]
	if expectedUrl != actualUrl {
		t.Errorf("expected %s, got %s", expectedUrl, actualUrl)
	}
	if itemExpected.Name != itemActual.Name {
		t.Errorf("expected %s, got %s", itemExpected.Name, itemActual.Name)
	}
	if itemExpected.Team != itemActual.Team {
		t.Errorf("expected %s, got %s", itemExpected.Team, itemActual.Team)
	}
}

func TestSetEnvVars(t *testing.T) {
	app := &App{Name: "teresa", Team: "luizalabs"}
	var testCases = []struct {
		evs  []*EnvVar
		want []*EnvVar
	}{
		{
			[]*EnvVar{
				{Key: "key2", Value: "value2"},
			},
			[]*EnvVar{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
		},
		{
			[]*EnvVar{
				{Key: "key1", Value: "new-value1"},
				{Key: "key2", Value: "value2"},
			},
			[]*EnvVar{
				{Key: "key1", Value: "new-value1"},
				{Key: "key2", Value: "value2"},
			},
		},
	}

	for _, tc := range testCases {
		app.EnvVars = []*EnvVar{{Key: "key1", Value: "value1"}}
		setEnvVars(app, tc.evs)
		if !reflect.DeepEqual(app.EnvVars, tc.want) {
			t.Errorf("expected %v, got %v", tc.want, app.EnvVars)
		}
	}
}

func TestUnsetEnvVars(t *testing.T) {
	app := &App{Name: "teresa", Team: "luizalabs"}
	var testCases = []struct {
		evs  []string
		want []*EnvVar
	}{
		{
			[]string{"key2"},
			[]*EnvVar{{Key: "key1", Value: "value1"}},
		},
		{
			[]string{"key1", "key2"},
			[]*EnvVar{},
		},
	}

	for _, tc := range testCases {
		app.EnvVars = []*EnvVar{
			{Key: "key1", Value: "value1"},
			{Key: "key2", Value: "value2"},
		}
		unsetEnvVars(app, tc.evs)
		if !reflect.DeepEqual(app.EnvVars, tc.want) {
			t.Errorf("expected %v, got %v", tc.want, app.EnvVars)
		}
	}
}

func TestSetSecretsOnApp(t *testing.T) {
	a := &App{Name: "teresa", Team: "luizalabs"}
	var testCases = []struct {
		actual  []string
		secrets []string
		want    []string
	}{
		{[]string{}, []string{"S1"}, []string{"S1"}},
		{[]string{"S1"}, []string{"S2"}, []string{"S1", "S2"}},
	}

	for _, tc := range testCases {
		a.Secrets = tc.actual
		setSecretsOnApp(a, tc.secrets)
		if !reflect.DeepEqual(a.Secrets, tc.want) {
			t.Errorf("expected %v, got %v", tc.want, a.Secrets)
		}
	}
}

func TestUnSetSecretsOnApp(t *testing.T) {
	a := &App{Name: "teresa", Team: "luizalabs"}
	var testCases = []struct {
		actual  []string
		secrets []string
		want    []string
	}{
		{[]string{"S1"}, []string{"S1"}, []string{}},
		{[]string{"S1", "S2"}, []string{"S2"}, []string{"S1"}},
	}

	for _, tc := range testCases {
		a.Secrets = tc.actual
		unsetSecretsOnApp(a, tc.secrets)
		if !reflect.DeepEqual(a.Secrets, tc.want) {
			t.Errorf("expected %v, got %v", tc.want, a.Secrets)
		}
	}
}

func TestSetSecretFileOnApp(t *testing.T) {
	a := &App{Name: "teresa", Team: "luizalabs"}
	var testCases = []struct {
		actual []string
		secret string
		want   []string
	}{
		{[]string{}, "S1", []string{"S1"}},
		{[]string{"S1"}, "S2", []string{"S1", "S2"}},
	}

	for _, tc := range testCases {
		a.SecretFiles = tc.actual
		setSecretFileOnApp(a, tc.secret)
		if !reflect.DeepEqual(a.SecretFiles, tc.want) {
			t.Errorf("expected %v, got %v", tc.want, a.SecretFiles)
		}
	}
}

func TestUnSetSecretFilesOnApp(t *testing.T) {
	a := &App{Name: "teresa", Team: "luizalabs"}
	var testCases = []struct {
		actual  []string
		secrets []string
		want    []string
	}{
		{[]string{"S1"}, []string{"S1"}, []string{}},
		{[]string{"S1", "S2"}, []string{"S2"}, []string{"S1"}},
		{[]string{"S1", "S2"}, []string{"S2", "S1"}, []string{}},
	}

	for _, tc := range testCases {
		a.SecretFiles = tc.actual
		unsetSecretFilesOnApp(a, tc.secrets)
		if !reflect.DeepEqual(a.SecretFiles, tc.want) {
			t.Errorf("expected %v, got %v", tc.want, a.SecretFiles)
		}
	}
}

func TestNewAutoscale(t *testing.T) {
	req := newAutoscaleRequest("teresa")
	as := newAutoscale(req)
	want := &Autoscale{
		CPUTargetUtilization: 10,
		Min:                  1,
		Max:                  2,
	}

	if !reflect.DeepEqual(as, want) {
		t.Errorf("got %v; want %v", as, want)
	}
}
