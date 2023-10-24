package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStripFileNameFromPath(t *testing.T) {
	t.Helper()
	t.Log(StripFileNameFromPath("messages/user79/image2291589349.png"))
}

func TestUsersDirectory(t *testing.T) {
	testTable := []struct {
		input       string
		expectedErr bool
		expectedId  int
	}{
		{
			input:       "/account/messages/",
			expectedErr: true,
			expectedId:  0,
		},
		{
			input:       "/account/user/",
			expectedErr: true,
			expectedId:  0,
		},
		{
			input:       "/account/messages/user78/",
			expectedErr: false,
			expectedId:  78,
		},
		{
			input:       "/account/messages/user78/image423297995.png",
			expectedErr: false,
			expectedId:  78,
		},
	}

	for _, testCase := range testTable {
		id, err := NumberOfUsersDirectory(testCase.input)
		if testCase.expectedErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, id, testCase.expectedId)
	}
}

func TestStripUrlFromFilePath(t *testing.T) {
	fullPath := "C:/Users/User/GolandProjects/Cataloguer/messages/user96/application4050530390.mp4"
	basePath := "C:\\Users\\User\\GolandProjects\\Cataloguer"
	t.Log(StripUrlFromFilePath(fullPath, basePath))
}
