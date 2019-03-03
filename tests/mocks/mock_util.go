package mocks

import (
	"testing"

	mk "github.com/golang/mock/gomock"
)

func WithTestLogging(mock *MockLogger, t *testing.T) *MockLogger {

	mock.EXPECT().
		Debug(mk.Any(), mk.Any()).
		Do(func(format string, args ...interface{}) {
			t.Logf(format, args...)
		}).
		AnyTimes()
	mock.EXPECT().
		Info(mk.Any(), mk.Any()).
		Do(func(format string, args ...interface{}) {
			t.Logf(format, args...)
		}).
		AnyTimes()

	return mock
}
