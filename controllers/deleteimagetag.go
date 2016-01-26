package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/eirka/eirka-libs/audit"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/redis"
	"github.com/eirka/eirka-libs/user"

	"github.com/eirka/eirka-admin/models"
)

// DeleteImageTagController will delete an image tag
func DeleteImageTagController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from user middleware
	userdata := c.MustGet("userdata").(user.User)

	// Initialize model struct
	m := &models.DeleteImageTagModel{
		Ib:    params[0],
		Image: params[1],
		Tag:   params[2],
	}

	// Check the record id and get further info
	err := m.Status()
	if err == e.ErrNotFound {
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("DeleteImageTagController.Status")
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("DeleteImageTagController.Status")
		return
	}

	// Delete data
	err = m.Delete()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("DeleteImageTagController.Delete")
		return
	}

	// Initialize cache handle
	cache := redis.RedisCache

	// Delete redis stuff
	tags_key := fmt.Sprintf("%s:%d", "tags", m.Ib)
	tag_key := fmt.Sprintf("%s:%d:%d", "tag", m.Ib, m.Tag)
	image_key := fmt.Sprintf("%s:%d", "image", m.Ib)

	err = cache.Delete(tags_key, tag_key, image_key)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("DeleteImageTagController.cache.Delete")
		return
	}

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": audit.AuditDeleteImageTag})

	// audit log
	audit := audit.Audit{
		User:   userdata.Id,
		Ib:     m.Ib,
		Ip:     c.ClientIP(),
		Action: audit.AuditDeleteImageTag,
		Info:   fmt.Sprintf("%d/%s", m.Image, m.Name),
	}

	// submit audit
	err = audit.Submit()
	if err != nil {
		c.Error(err).SetMeta("DeleteImageTagController.audit.Submit")
	}

	return

}
