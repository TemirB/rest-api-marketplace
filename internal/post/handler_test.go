package post

import (
	"net/url"
	"testing"
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

// // TODO: Add test cases for GetPosts and UpdatePost functions
// func Test_CreatePost(t *testing.T) {
// 	testCases := []struct {
// 		name string

// 		method string
// 		post   Post

// 		expected Post
// 	}{
// 		{
// 			name: "1. Valid_Post",
// 			post: Post{
// 				Title:       "Test Post",
// 				Description: "This is a test post.",
// 				Price:       10.0,
// 				ImageURL:    "https://example.com/image.jpg",
// 				Owner:       "testuser",
// 			},

// 			expected: Post{
// 				Title:       "Test Post",
// 				Description: "This is a test post.",
// 				Price:       10.0,
// 				ImageURL:    "https://example.com/image.jpg",
// 				Owner:       "testuser",
// 				ID:          1,
// 			},
// 		},
// 	}
// }
