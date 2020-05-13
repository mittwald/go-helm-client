package mockhelmclient

import (
	"testing"

	"github.com/golang/mock/gomock"
)

func TestUpdateChartRepos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockClient(ctrl)

	t.Run("UpdateChartRepos", func(t *testing.T) {
		mockClient.EXPECT().UpdateChartRepos()
		err := mockClient.UpdateChartRepos()
		if err != nil {
			panic(err)
		}
	})
}
