// Copyright 2020 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package worker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	mocklib "github.com/stretchr/testify/mock"

	"github.com/mendersoftware/workflows/model"
	"github.com/mendersoftware/workflows/store/mock"
)

func TestProcessJobCLI(t *testing.T) {
	ctx := context.Background()
	dataStore := mock.NewDataStore()

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
				CLI: &model.CLITask{
					Command: []string{
						"echo",
						"TEST",
					},
				},
			},
		},
	}

	job := &model.Job{
		WorkflowName: workflow.Name,
		Status:       model.StatusPending,
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
	).Return(workflow, nil)

	dataStore.On("AquireJob",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
	).Return(job, nil)

	dataStore.On("UpdateJobStatus",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		model.StatusDone,
	).Return(nil)

	dataStore.On("UpdateJobAddResult",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		mocklib.MatchedBy(
			func(taskResult *model.TaskResult) bool {
				assert.True(t, taskResult.Success)
				assert.Equal(t, workflow.Tasks[0].CLI.Command, taskResult.CLI.Command)
				assert.Equal(t, "TEST\n", taskResult.CLI.Output)
				assert.Equal(t, "", taskResult.CLI.Error)
				assert.Equal(t, 0, taskResult.CLI.ExitCode)

				return true
			}),
	).Return(nil)

	err := processJob(ctx, job, dataStore)

	assert.Nil(t, err)

	dataStore.AssertExpectations(t)
}

func TestProcessJobCLIWrongExitCode(t *testing.T) {
	ctx := context.Background()
	dataStore := mock.NewDataStore()

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
				CLI: &model.CLITask{
					Command: []string{
						"bash",
						"-c",
						"exit 10",
					},
				},
			},
			{
				Name: "task_2",
				Type: model.TaskTypeHTTP,
				HTTP: &model.HTTPTask{
					URI:    "http://localhost",
					Method: "GET",
					Headers: map[string]string{
						"X-Header": "Value",
					},
					StatusCodes: []int{
						200,
						201,
					},
				},
			},
		},
	}

	job := &model.Job{
		WorkflowName: workflow.Name,
		Status:       model.StatusPending,
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
	).Return(workflow, nil)

	dataStore.On("AquireJob",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
	).Return(job, nil)

	dataStore.On("UpdateJobStatus",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		model.StatusFailure,
	).Return(nil)

	dataStore.On("UpdateJobAddResult",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		mocklib.MatchedBy(
			func(taskResult *model.TaskResult) bool {
				assert.False(t, taskResult.Success)
				assert.Equal(t, workflow.Tasks[0].CLI.Command, taskResult.CLI.Command)
				assert.Equal(t, "", taskResult.CLI.Output)
				assert.Equal(t, "", taskResult.CLI.Error)
				assert.Equal(t, 10, taskResult.CLI.ExitCode)

				return true
			}),
	).Return(nil)

	err := processJob(ctx, job, dataStore)
	assert.Nil(t, err)

	dataStore.AssertExpectations(t)
}

func TestProcessJobCLTimeOut(t *testing.T) {
	ctx := context.Background()
	dataStore := mock.NewDataStore()

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
				CLI: &model.CLITask{
					Command: []string{
						"bash",
						"-c",
						"sleep 10",
					},
					ExecutionTimeOut: 2,
				},
			},
		},
	}

	job := &model.Job{
		WorkflowName: workflow.Name,
		Status:       model.StatusPending,
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
	).Return(workflow, nil)

	dataStore.On("AquireJob",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
	).Return(job, nil)

	dataStore.On("UpdateJobStatus",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		model.StatusFailure,
	).Return(nil)

	dataStore.On("UpdateJobAddResult",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		mocklib.MatchedBy(
			func(taskResult *model.TaskResult) bool {
				assert.False(t, taskResult.Success)
				assert.Equal(t, workflow.Tasks[0].CLI.Command, taskResult.CLI.Command)
				assert.Equal(t, "", taskResult.CLI.Output)
				assert.Equal(t, "", taskResult.CLI.Error)
				assert.Equal(t, -1, taskResult.CLI.ExitCode)

				return true
			}),
	).Return(nil)

	err := processJob(ctx, job, dataStore)
	assert.Nil(t, err)

	dataStore.AssertExpectations(t)
}

func TestProcessJobCLIFailedIncompatibleDefinition(t *testing.T) {
	ctx := context.Background()
	dataStore := mock.NewDataStore()

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeCLI,
			},
		},
	}

	job := &model.Job{
		WorkflowName: workflow.Name,
		Status:       model.StatusPending,
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
	).Return(workflow, nil)

	dataStore.On("AquireJob",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
	).Return(job, nil)

	dataStore.On("UpdateJobStatus",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		model.StatusFailure,
	).Return(nil)

	err := processJob(ctx, job, dataStore)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "Error: Task definition incompatible with specified type (cli)")

	dataStore.AssertExpectations(t)
}
