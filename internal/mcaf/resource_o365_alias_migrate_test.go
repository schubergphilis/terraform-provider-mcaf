package mcaf

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

func TestResourceO365AliasStateUpgradeV0(t *testing.T) {

	groupName := acctest.RandString(8)
	groupUuid := uuid.New().String()

	expected := testResourceO365AliasStateDataV1(groupUuid)
	actual, err := resourceO365AliasStateUpgradeV0(context.Background(), testResourceO365AliasStateDataV0(groupName, groupUuid), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}

func testResourceO365AliasStateDataV0(name, uuid string) map[string]interface{} {
	return map[string]interface{}{
		"group_id": fmt.Sprintf("%s_%s", name, uuid),
	}
}

func testResourceO365AliasStateDataV1(uuid string) map[string]interface{} {
	return map[string]interface{}{
		"group_id": uuid,
	}
}
