package post

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/TemirB/rest-api-marketplace/internal/middleware"
)

func Test_setFilter(t *testing.T) {
	testCases := []struct {
		name string

		q     url.Values
		owner string

		expected FilterParams
	}{
		{
			name:  "1. No_Filter_Params",
			q:     url.Values{},
			owner: "testuser",

			expected: FilterParams{
				MinPrice: 0,
				MaxPrice: -1,
				Owner:    "testuser",
			},
		},
		{
			name: "2. Valid_Filter_Params",
			q: url.Values{
				"min_price": {"10"},
				"max_price": {"20"},
			},
			owner: "testuser",

			expected: FilterParams{
				MinPrice: 10,
				MaxPrice: 20,
				Owner:    "testuser",
			},
		},
		{
			name: "3. Invalid_Max_Price_Param",
			q: url.Values{
				"min_price": {"10"},
				"max_price": {"-20"},
			},
			owner: "testuser",

			expected: FilterParams{
				MinPrice: 10,
				MaxPrice: -1,
				Owner:    "testuser",
			},
		},
		{
			name: "4. Invalid_Min_Price_Param",
			q: url.Values{
				"min_price": {"-10"},
				"max_price": {"20"},
			},
			owner: "testuser",

			expected: FilterParams{
				MinPrice: 0,
				MaxPrice: 20,
				Owner:    "testuser",
			},
		},
		{
			name: "5. Both_Price_Params_Are_Invalid",
			q: url.Values{
				"min_price": {"-10"},
				"max_price": {"-20"},
			},
			owner: "testuser",

			expected: FilterParams{
				MinPrice: 0,
				MaxPrice: -1,
				Owner:    "testuser",
			},
		},
		{
			name: "6. Both_Price_Params_Are_Valid_but_Min_Is_Greater_Than_Max",
			q: url.Values{
				"min_price": {"20"},
				"max_price": {"10"},
			},
			owner: "testuser",

			expected: FilterParams{
				MinPrice: 10,
				MaxPrice: 20,
				Owner:    "testuser",
			},
		},
		{
			name:  "7. Owner_Filter_Param",
			q:     url.Values{"owner": []string{"bob"}},
			owner: "alice",

			expected: FilterParams{
				MinPrice: 0,
				MaxPrice: -1,
				Owner:    "alice",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := setFilter(tc.q, tc.owner)
			if actual.MinPrice != tc.expected.MinPrice || actual.MaxPrice != tc.expected.MaxPrice || actual.Owner != tc.expected.Owner {
				t.Errorf("Expected %+v, got %+v", tc.expected, actual)
			}
		})
	}
}

func Test_setSort(t *testing.T) {
	testCases := []struct {
		name string

		q url.Values

		expected SortParams
	}{
		{
			name: "1. No_Sort_Param",
			q:    url.Values{},

			expected: SortParams{
				Field:     "created_at",
				Direction: "DESC",
			},
		},
		{
			name: "2. Valid_Sort_Param",
			q: url.Values{
				"sort_by": {"created_at"},
			},

			expected: SortParams{
				Field:     "created_at",
				Direction: "DESC",
			},
		},
		{
			name: "3. Invalid_Sort_Param",
			q: url.Values{
				"sort_by": {"invalid_sort_param"},
			},

			expected: SortParams{
				Field:     "created_at",
				Direction: "DESC",
			},
		},
		{
			name: "4. Valid_Sort_Param_In_Ascending_Order",
			q: url.Values{
				"sort_by": {"price"},
				"order":   {"asc"},
			},

			expected: SortParams{
				Field:     "price",
				Direction: "ASC",
			},
		},
		{
			name: "5. Valid_Sort_Param_In_Descending_Order",
			q: url.Values{
				"sort_by": {"price"},
				"order":   {"desc"},
			},

			expected: SortParams{
				Field:     "price",
				Direction: "DESC",
			},
		},
		{
			name: "6. Invalid_Order_Param",
			q: url.Values{
				"sort_by": {"price"},
				"order":   {"invalid_order_param"},
			},

			expected: SortParams{
				Field:     "price",
				Direction: "DESC",
			},
		},
		{
			name: "7. Both_Sort_And_Order_Params_Are_Invalid",
			q: url.Values{
				"sort_by": {"invalid_sort_param"},
				"order":   {"invalid_order_param"},
			},

			expected: SortParams{
				Field:     "created_at",
				Direction: "DESC",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := setSort(tc.q)
			if actual.Field != tc.expected.Field || actual.Direction != tc.expected.Direction {
				t.Errorf("Expected %+v, got %+v", tc.expected, actual)
			}
		})
	}
}

func TestHandler_GetPosts_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := NewMockservice(ctrl)
	handler := NewHandler(mockService, zap.NewNop())

	posts := []*Post{
		{ID: 1, Title: "First", Price: 50, Owner: "alice", IsOwner: true},
		{ID: 2, Title: "Second", Price: 150, Owner: "bob", IsOwner: false},
	}
	mockService.EXPECT().GetPosts(gomock.Any(), gomock.Any()).Return(posts, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/feed", nil)
	ctx := context.WithValue(req.Context(), middleware.CtxUser, "alice")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetPosts(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var respPosts []Post
	err := json.Unmarshal(rr.Body.Bytes(), &respPosts)
	assert.NoError(t, err)
	assert.Len(t, respPosts, 2)

	for _, p := range respPosts {
		if p.Owner == "alice" {
			assert.True(t, p.IsOwner)
		} else {
			assert.False(t, p.IsOwner)
		}
	}
}

func TestHandler_GetPosts_WrongMethod(t *testing.T) {
	handler := NewHandler(nil, zap.NewNop())

	req := httptest.NewRequest(http.MethodDelete, "/posts/feed", nil)
	rr := httptest.NewRecorder()

	handler.GetPosts(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandler_GetPosts_ServiceErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := NewMockservice(ctrl)
	handler := NewHandler(mockService, zap.NewNop())
	mockService.EXPECT().GetPosts(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))

	req := httptest.NewRequest(http.MethodGet, "/posts/feed", nil)
	ctx := context.WithValue(req.Context(), middleware.CtxUser, "alice")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetPosts(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandler_GetPosts_NoAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := NewMockservice(ctrl)
	handler := NewHandler(mockService, zap.NewNop())
	posts := []*Post{
		{ID: 1, Title: "First", Owner: "alice", Price: 50, IsOwner: false},
		{ID: 2, Title: "Second", Owner: "bob", Price: 150, IsOwner: false},
	}
	mockService.EXPECT().GetPosts(gomock.Any(), gomock.Any()).Return(posts, nil)
	req := httptest.NewRequest(http.MethodGet, "/posts/feed", nil)
	rr := httptest.NewRecorder()
	handler.GetPosts(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	var respPosts []Post
	_ = json.Unmarshal(rr.Body.Bytes(), &respPosts)
	assert.Len(t, respPosts, 2)
	for _, p := range respPosts {
		assert.False(t, p.IsOwner)
	}
}

func TestHandler_GetPosts_FilterByOtherOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := NewMockservice(ctrl)
	handler := NewHandler(mockService, zap.NewNop())
	posts := []*Post{
		{ID: 1, Title: "Bob1", Owner: "bob", Price: 10, IsOwner: false},
		{ID: 2, Title: "Bob2", Owner: "bob", Price: 20, IsOwner: false},
	}
	mockService.EXPECT().
		GetPosts(gomock.Any(), gomock.Any()).
		Return(posts, nil)
	req := httptest.NewRequest(http.MethodGet, "/posts/feed?owner=bob", nil)
	ctx := context.WithValue(req.Context(), middleware.CtxUser, "alice")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.GetPosts(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	var respPosts []Post
	_ = json.Unmarshal(rr.Body.Bytes(), &respPosts)
	assert.Len(t, respPosts, 2)
	for _, p := range respPosts {
		assert.False(t, p.IsOwner)
	}
}

func TestHandler_CreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := []struct {
		name string

		method     string
		url        string
		body       []byte
		setupMocks func(mockService *Mockservice)

		unauthorized bool
		expectedCode int
	}{
		{
			name:   "1. Valid_Create_Post_Request",
			method: http.MethodPost,
			url:    "/posts",
			body:   []byte(`{"title": "New Post", "price": 100, "owner": "alice", "description": "This is a new post", "image_url": "new_post.jpg"}`),
			setupMocks: func(mockService *Mockservice) {
				post := &Post{Title: "New Post", Price: 100, Owner: "alice", Description: "This is a new post", ImageURL: "new_post.jpg"}
				mockService.EXPECT().CreatePost(post).Return(post, nil)
			},

			expectedCode: http.StatusCreated,
		},
		{
			name:   "2. Invalid_Create_Post_Wrong_Method",
			method: http.MethodDelete,
			url:    "/posts",
			body:   []byte(`{"price": 100, "owner": "alice", "description": "This is a new post", "image_url": "new_post.jpg"}`),
			setupMocks: func(mockService *Mockservice) {
				// No-op
			},

			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:   "3. Invalid_Create_Post_Unautharized",
			method: http.MethodPost,
			url:    "/posts",
			body:   []byte(`{"price": 100, "owner": "alice", "description": "This is a new post", "image_url": "new_post.jpg"}`),
			setupMocks: func(mockService *Mockservice) {
				// No-op
			},

			unauthorized: true,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:   "4. Invalid_JSON_Body",
			method: http.MethodPost,
			url:    "/posts",
			body:   []byte(`{"price": 100, "owner": "alice", "description": "This is a new post", "image_url": "new_post.jpg",`),
			setupMocks: func(mockService *Mockservice) {
				// No-op
			},

			expectedCode: http.StatusBadRequest,
		},
		{
			name:   "5. Service_Error",
			method: http.MethodPost,
			url:    "/posts",
			body:   []byte(`{"price": 100, "owner": "alice", "description": "This is a new post", "image_url": "new_post.jpg"}`),
			setupMocks: func(mockService *Mockservice) {
				mockService.EXPECT().CreatePost(gomock.Any()).Return(nil, errors.New("service error"))
			},

			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewMockservice(ctrl)
			tc.setupMocks(service)
			handler := NewHandler(service, zap.NewNop())
			req, _ := http.NewRequest(tc.method, tc.url, bytes.NewBuffer(tc.body))
			if !tc.unauthorized {
				ctx := context.WithValue(req.Context(), middleware.CtxUser, "alice")
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler.CreatePost(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
		})
	}
}
