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

// PurgeThreadController will remove deleted files and rows
func PurgeThreadController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from user middleware
	userdata := c.MustGet("userdata").(user.User)

	// Initialize model struct
	m := &models.PurgeThreadModel{
		Ib: params[0],
		Id: params[1],
	}

	// Check the record id and get further info
	err := m.Status()
	if err == e.ErrNotFound {
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("PurgeThreadController.Status")
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("PurgeThreadController.Status")
		return
	}

	// Delete data
	err = m.Delete()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("PurgeThreadController.Delete")
		return
	}

	// Initialize cache handle
	cache := redis.RedisCache

	// Delete redis stuff
	index_key := fmt.Sprintf("%s:%d", "index", m.Ib)
	directory_key := fmt.Sprintf("%s:%d", "directory", m.Ib)
	thread_key := fmt.Sprintf("%s:%d:%d", "thread", m.Ib, m.Id)
	post_key := fmt.Sprintf("%s:%d:%d", "post", m.Ib, m.Id)
	tags_key := fmt.Sprintf("%s:%d", "tags", m.Ib)
	image_key := fmt.Sprintf("%s:%d", "image", m.Ib)
	new_key := fmt.Sprintf("%s:%d", "new", m.Ib)
	popular_key := fmt.Sprintf("%s:%d", "popular", m.Ib)
	favorited_key := fmt.Sprintf("%s:%d", "favorited", m.Ib)

	err = cache.Delete(index_key, directory_key, thread_key, post_key, tags_key, image_key, new_key, popular_key, favorited_key)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("PurgeThreadController.cache.Delete")
		return
	}

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": audit.AuditPurgeThread})

	// audit log
	audit := audit.Audit{
		User:   userdata.Id,
		Ib:     m.Ib,
		Ip:     c.ClientIP(),
		Action: audit.AuditPurgeThread,
		Info:   fmt.Sprintf("%s", m.Name),
	}

	// submit audit
	err = audit.Submit()
	if err != nil {
		c.Error(err).SetMeta("PurgeThreadController.audit.Submit")
	}

	return

}
