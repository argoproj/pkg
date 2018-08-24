package errors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestIsRequestEntityTooLargeErr test
func TestIsRequestEntityTooLargeErr(t *testing.T) {
	assert.False(t, IsRequestEntityTooLargeErr(nil))

	var err error
	err = &apierr.StatusError{metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    http.StatusRequestEntityTooLarge,
		Reason:  "",
		Message: "",
	}}
	assert.True(t, IsRequestEntityTooLargeErr(err))

	err = &apierr.StatusError{metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    http.StatusInternalServerError,
		Reason:  "",
		Message: "etcdserver: request is too large",
	}}
	assert.True(t, IsRequestEntityTooLargeErr(err))

	err = &apierr.StatusError{metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    http.StatusInternalServerError,
		Reason:  "",
		Message: "",
	}}
	assert.False(t, IsRequestEntityTooLargeErr(err))

}
