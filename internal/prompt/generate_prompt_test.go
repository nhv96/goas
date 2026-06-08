package prompt

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestCreateGeneratePrompt(t *testing.T) {
	rdr, err := CreateGeneratePrompt("model_a", "mess", "sys_mess", true, false)
	assert.Nil(t, err)

	actualBytes, err := io.ReadAll(rdr)
	assert.Nil(t, err)

	actual := &GeneratePrompt{}
	err = json.Unmarshal(actualBytes, actual)
	assert.Nil(t, err)

	expectedJSON := []byte(`{"model":"model_a","prompt":"mess","system":"sys_mess","think":true,"stream":false}`)
	expect := &GeneratePrompt{}
	err = json.Unmarshal(expectedJSON, expect)
	assert.Nil(t, err)

	if diff := cmp.Diff(expect, actual); diff != "" {
		t.Error(diff)
	}
}
