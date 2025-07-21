package post

import (
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_CreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := []struct {
		name string

		post       *Post
		setupMocks func(storage *Mockstorage, post *Post)

		expectedError error
	}{
		{
			name: "Valid post",

			post: &Post{
				Title:       "Test post",
				Description: "This is a test post",
				Price:       100.50,
				ImageURL:    "https://example.com/image.jpg",
			},

			setupMocks: func(storage *Mockstorage, post *Post) {
				storage.EXPECT().Create(post).Return(nil)
			},
		},
		{
			name: "Missing title",

			post: &Post{
				Description: "This is a test post",
				Price:       100.50,
				ImageURL:    "https://example.com/image.jpg",
			},
			setupMocks: func(storage *Mockstorage, post *Post) {},

			expectedError: ErrMissingFields,
		},
		{
			name: "Too long title",
			post: &Post{
				Title: "Отче наш, Иже еси на небесе́х!Да святится имя Твое, да прии́дет Царствие Твое, Да будет воля Твоя, яко на небеси́ и на земли́. Хлеб наш насущный да́ждь нам дне́сь; И оста́ви нам до́лги наша, якоже и мы оставляем должнико́м нашим; И не введи нас во искушение, но изба́ви нас от лукаваго. Яко Твое есть Царство и сила, и слава, Отца, и Сына, и Святаго Духа, ныне и присно, и во веки веков. Аминь",
			},
			setupMocks: func(storage *Mockstorage, post *Post) {},

			expectedError: ErrTooLong,
		},
		{
			name: "Negative price",

			post: &Post{
				Title:       "Test post",
				Description: "This is a test post",
				Price:       -100.50,
				ImageURL:    "https://example.com/image.jpg",
			},
			setupMocks: func(storage *Mockstorage, post *Post) {},

			expectedError: ErrNegativePrice,
		},
		{
			name: "Missing image URL",

			post: &Post{
				Title:       "Test post",
				Description: "This is a test post",
				Price:       100.50,
			},
			setupMocks: func(storage *Mockstorage, post *Post) {},

			expectedError: ErrRequiredURL,
		},
		{
			name: "Storage error",

			post: &Post{
				Title:       "Test post",
				Description: "This is a test post",
				Price:       100.50,
				ImageURL:    "https://example.com/image.jpg",
			},
			setupMocks: func(storage *Mockstorage, post *Post) {
				storage.EXPECT().Create(post).Return(fmt.Errorf("storage error"))
			},

			expectedError: fmt.Errorf("storage error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMockstorage(ctrl)
			service := NewService(storage, zap.NewNop())
			tc.setupMocks(storage, tc.post)

			actual, err := service.CreatePost(tc.post)
			if err != nil && tc.expectedError == nil {
				assert.Equal(t, tc.post, actual)
			} else if err == nil && tc.expectedError != nil {
				t.Errorf("Expected error: %v, got: nil", tc.expectedError)
			}
		})
	}
}
