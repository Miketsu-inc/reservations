package validate_test

import (
	"testing"

	. "github.com/miketsu-inc/reservations/backend/pkg/validate"
	"github.com/stretchr/testify/assert"
)

func TestMerchantNameToUrlName(t *testing.T) {
	assert := assert.New(t)

	testcases := []string{"lksgd*sdhf@hd#Kn8    88dlas", "áasdÜioű. --asdl7v 79(ÉÁ' asd) oav87pz", "Űő cxm,n76  jakl(*,.1)?"}
	results := []string{"lksgd-sdhf-hd-Kn8-88dlas", "aasdUiou-asdl7v-79-EA-asd-oav87pz", "Uo-cxm-n76-jakl-1"}

	for i, s := range testcases {
		result, err := MerchantNameToUrlName(s)
		assert.Nil(err)

		assert.Equal(result, results[i])
	}

	_, err := MerchantNameToUrlName("*(^&)")
	assert.NotNil(err)
}
