package helmclient

import (
	"testing"

	"github.com/golang/mock/gomock"
)

func TestUpdateChartRepos(t *testing.T) {
	var e error
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockClient(ctrl)

	mockClient.EXPECT().
		UpdateChartRepos().
		Return(e).
		Times(1).
		Do(mockClient.UpdateChartRepos())
}
