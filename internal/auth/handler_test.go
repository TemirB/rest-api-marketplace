package auth

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := []struct {
		testID string
		name   string

		method  string
		url     string
		reqBody []byte

		setupMocks func(ctrl *gomock.Controller) *Handler

		expectedStatus int
	}{
		{
			testID: "1",
			name:   "Register_Success",

			method:  http.MethodPost,
			url:     "/register",
			reqBody: []byte(`{"login": "testUser", "password": "testPassword"}`),

			setupMocks: func(ctrl *gomock.Controller) *Handler {
				mockService := NewMockservice(ctrl)
				mockService.EXPECT().Register("testUser", "testPassword").Return(nil)

				return NewHandler(mockService, zap.NewNop())
			},

			expectedStatus: http.StatusCreated,
		},
		{
			testID: "2",
			name:   "Register_Error_Wrong_Method",

			method:  http.MethodDelete,
			url:     "/register",
			reqBody: []byte(`{"login": "testUser", "password": "testPassword"}`),

			setupMocks: func(ctrl *gomock.Controller) *Handler {
				return NewHandler(nil, zap.NewNop())
			},

			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			testID: "3",
			name:   "Register_Error_Decoding_Error",

			method:  http.MethodPost,
			url:     "/register",
			reqBody: []byte(``),

			setupMocks: func(ctrl *gomock.Controller) *Handler {
				return NewHandler(nil, zap.NewNop())
			},

			expectedStatus: http.StatusBadRequest,
		},
		{
			testID: "4",
			name:   "Register_Error_Service_Error",

			method:  http.MethodPost,
			url:     "/register",
			reqBody: []byte(`{"login": "testUser", "password": "testPassword"}`),

			setupMocks: func(ctrl *gomock.Controller) *Handler {
				mockService := NewMockservice(ctrl)
				mockService.EXPECT().Register("testUser", "testPassword").Return(errors.New("service error"))

				return NewHandler(mockService, zap.NewNop())
			},

			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := tc.setupMocks(ctrl)

			req, err := http.NewRequest(tc.method, tc.url, bytes.NewBuffer(tc.reqBody))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.Register(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := []struct {
		name string

		method  string
		url     string
		reqBody []byte

		setupMocks func(ctrl *gomock.Controller) *Handler

		expectedStatus int
	}{
		{
			name: "Login_Success",

			method:  http.MethodPost,
			url:     "/login",
			reqBody: []byte(`{"login": "testUser", "password": "testPassword"}`),

			setupMocks: func(ctrl *gomock.Controller) *Handler {
				mockService := NewMockservice(ctrl)
				mockService.EXPECT().Login("testUser", "testPassword").Return("testToken", nil)

				return NewHandler(mockService, zap.NewNop())
			},

			expectedStatus: http.StatusOK,
		},
		{
			name: "Login_Error_Wrong_Method",

			method:  http.MethodDelete,
			url:     "/login",
			reqBody: []byte(`{"login": "testUser", "password": "testPassword"}`),

			setupMocks: func(ctrl *gomock.Controller) *Handler {
				return NewHandler(nil, zap.NewNop())
			},

			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name: "Login_Error_Decoding_Error",

			method:  http.MethodPost,
			url:     "/login",
			reqBody: []byte(``),

			setupMocks: func(ctrl *gomock.Controller) *Handler {
				return NewHandler(nil, zap.NewNop())
			},

			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Login_Error_Service_Error",

			method:  http.MethodPost,
			url:     "/login",
			reqBody: []byte(`{"login": "testUser", "password": "testPassword"}`),

			setupMocks: func(ctrl *gomock.Controller) *Handler {
				mockService := NewMockservice(ctrl)
				mockService.EXPECT().Login("testUser", "testPassword").Return("", errors.New("service error"))

				return NewHandler(mockService, zap.NewNop())
			},

			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := tc.setupMocks(ctrl)

			req, err := http.NewRequest(tc.method, tc.url, bytes.NewBuffer(tc.reqBody))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.Login(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}
