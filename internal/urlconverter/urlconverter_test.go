package urlconverter_test

import (
	"testing"

	"github.com/dnfd/url_shortener/internal/urlconverter"
	"github.com/stretchr/testify/assert"
)

func TestConverter_RestoreID(t *testing.T) {
	var id int64 = 42

	url := urlconverter.IDToURL(id)
	restored, err := urlconverter.URLToID(url)

	assert.NoError(t, err)
	assert.Equal(t, id, restored)
}
