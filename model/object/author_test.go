package object_test

import (
	"testing"
	"time"

	"github.com/mizuho-u/got/model/object"
)

func TestAuthorString(t *testing.T) {

	tt := []struct {
		name   string
		email  string
		now    func() time.Time
		expect string
	}{
		{
			name:   "James Coglan",
			email:  "james@jcoglan.com",
			now:    func() time.Time { return time.Unix(1511204319, 0).UTC() },
			expect: "James Coglan <james@jcoglan.com> 1511204319 +0000",
		},
		{
			name:  "Mizuho Ueda",
			email: "mi_ueda@m-u.dev",
			now: func() time.Time {

				jst, _ := time.LoadLocation("Asia/Tokyo")
				return time.Unix(1677139170, 0).In(jst)
			},
			expect: "Mizuho Ueda <mi_ueda@m-u.dev> 1677139170 +0900",
		},
	}

	for _, tc := range tt {

		author := object.NewAuthor(tc.name, tc.email, tc.now())

		if author.String() != tc.expect {
			t.Errorf("author string not match. expect %s got %s", tc.expect, author)
		}

	}

}
