package chaincode_test

import (
	"testing"

	"github.com/bestbeforetoday/fabric-admin/pkg/chaincode"
	"github.com/stretchr/testify/require"
)

func TestLifecycle(t *testing.T) {
	t.Run("NewLifeCycle", func(t *testing.T) {
		actual := chaincode.NewLifecycle(nil)
		require.NotNil(t, actual)
	})
}
