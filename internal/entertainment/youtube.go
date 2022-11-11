package entertainment

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/eternal-flame-AD/yoake/internal/auth"
	"github.com/eternal-flame-AD/yoake/internal/db"
	"github.com/eternal-flame-AD/yoake/internal/echoerror"
	"github.com/labstack/echo/v4"
)

type YoutubeVideoStore struct {
	VideoID  string                 `json:"video_id" form:"video_id" query:"video_id"`
	Meta     *YoutubeVideoEmbedMeta `json:"meta,omitempty"`
	Tags     []string               `json:"tags"`
	Category string                 `param:"category" json:"category" form:"category" query:"category"`

	Comment string `json:"comment" form:"comment"`
}

type YoutubeVideoEmbedMeta struct {
	Title           string `json:"title"`
	AuthorName      string `json:"author_name"`
	AuthorURL       string `json:"author_url"`
	Type            string `json:"type"`
	ProviderName    string `json:"provider_name"`
	ProviderURL     string `json:"provider_url"`
	ThumbnailURL    string `json:"thumbnail_url"`
	ThumbnailWidth  int    `json:"thumbnail_width"`
	ThumbnailHeight int    `json:"thumbnail_height"`
	Html            string `json:"html"`
	Version         string `json:"version"`
	Height          int    `json:"height"`
	Width           int    `json:"width"`
}

func GetYoutubeVideoInfo(videoUrl string) (info *YoutubeVideoEmbedMeta, err error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get("https://www.youtube.com/oembed?format=json&url=https://www.youtube.com/watch?v=" + url.QueryEscape(videoUrl))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	info = new(YoutubeVideoEmbedMeta)
	err = json.NewDecoder(resp.Body).Decode(info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

type YoutubeCategoryStore struct {
	ID          string `json:"id" form:"id"`
	DisplayName string `json:"display_name" form:"display_name"`
}

type YoutubeTagStore struct {
	ID          string `json:"id" form:"id"`
	DisplayName string `json:"display_name" form:"display_name"`
}

type YoutubeVideoDBTxn struct {
	txn db.DBTxn
}

func newYoutubeDBTxn(txn db.DBTxn) *YoutubeVideoDBTxn {
	return &YoutubeVideoDBTxn{
		txn: txn,
	}
}

func (t *YoutubeVideoDBTxn) GetCategories() (categories []YoutubeCategoryStore, err error) {
	err = db.GetJSON(t.txn, []byte("youtube_categories"), &categories)
	return
}

func (t *YoutubeVideoDBTxn) SetCategories(categories []YoutubeCategoryStore) (err error) {
	return db.SetJSON(t.txn, []byte("youtube_categories"), categories)
}

func (t *YoutubeVideoDBTxn) DeleteCategory(category string) (err error) {
	categories, err := t.GetCategories()
	if err != nil {
		return err
	}
	newCategories := make([]YoutubeCategoryStore, 0, len(categories))
	for _, categoryS := range categories {
		if categoryS.ID != category {
			newCategories = append(newCategories, categoryS)
		}
	}
	if err = t.SetCategories(newCategories); err != nil {
		return err
	}

	if err := t.txn.Delete([]byte("youtube_category:" + category + "_tags")); err != nil && !db.IsNotFound(err) {
		return err
	}
	if err := t.txn.Delete([]byte("youtube_category:" + category + "_videos")); err != nil && !db.IsNotFound(err) {
		return err
	}
	return nil
}

func (t *YoutubeVideoDBTxn) GetTags(category string) (tags []YoutubeTagStore, err error) {
	categories, err := t.GetCategories()
	if err != nil {
		return nil, err
	}
	for _, categoryS := range categories {
		if categoryS.ID == category {
			err = db.GetJSON(t.txn, []byte("youtube_category:"+category+"_tags"), &tags)
			return
		}
	}
	return nil, echoerror.NewHttp(404, fmt.Errorf("category not found"))
}

func (t *YoutubeVideoDBTxn) SetTags(category string, tags []YoutubeTagStore) (err error) {
	categories, err := t.GetCategories()
	if err != nil {
		return err
	}
	for _, categoryS := range categories {
		if categoryS.ID == category {

			return db.SetJSON(t.txn, []byte("youtube_category:"+category+"_tags"), tags)
		}
	}
	return echoerror.NewHttp(404, fmt.Errorf("category not found"))
}

func (t *YoutubeVideoDBTxn) GetVideos(category string, tags []string) (videos []YoutubeVideoStore, err error) {
	videos = make([]YoutubeVideoStore, 0, 16)
	categories, err := t.GetCategories()
	if err != nil {
		return nil, err
	}
	for _, categoryS := range categories {
		if categoryS.ID == category {
			tagsAvail, err := t.GetTags(category)
			if err != nil {
				return nil, err
			}
			var tagSelected []string
			for _, tag := range tags {
				for _, tagAvail := range tagsAvail {
					if tagAvail.ID == tag {
						tagSelected = append(tagSelected, tag)
						break
					}
				}
			}

			var videosS []YoutubeVideoStore
			if err = db.GetJSON(t.txn, []byte("youtube_category:"+category+"_videos"), &videosS); err != nil {
				return nil, err
			}
			if len(tagSelected) == 0 {
				return videosS, nil
			}
			for _, video := range videosS {
			matchtag:
				for _, tagA := range tagSelected {
					for _, tag := range video.Tags {
						if tagA == tag {
							videos = append(videos, video)
							break matchtag
						}
					}
				}
			}
		}
	}
	return videos, nil
}

func (t *YoutubeVideoDBTxn) SetVideos(category string, videos []YoutubeVideoStore) (err error) {
	categories, err := t.GetCategories()
	if err != nil {
		return err
	}
	for _, categoryS := range categories {
		if categoryS.ID == category {
			existingTags, err := t.GetTags(category)
			if err != nil {
				return err
			}
			tagsUsed := make(map[string]YoutubeTagStore)
			for _, video := range videos {
				for _, tag := range video.Tags {
					found := false
					for _, existingTag := range existingTags {
						if existingTag.ID == tag {
							tagsUsed[tag] = existingTag
							found = true
							break
						}
					}
					if !found {
						return echoerror.NewHttp(400, fmt.Errorf("tag %s not found", tag))
					}
				}
			}
			tagsUsedList := make([]YoutubeTagStore, 0, len(tagsUsed))
			for _, tag := range tagsUsed {
				tagsUsedList = append(tagsUsedList, tag)
			}
			if err = t.SetTags(category, tagsUsedList); err != nil {
				return err
			}
			return db.SetJSON(t.txn, []byte("youtube_category:"+category+"_videos"), videos)
		}
	}
	return echoerror.NewHttp(404, fmt.Errorf("category not found"))
}

func registerYoutube(g *echo.Group, database db.DB) {
	g.GET("/categories", func(c echo.Context) error {
		txn := newYoutubeDBTxn(database.NewTransaction(false))
		defer txn.txn.Discard()

		categories, err := txn.GetCategories()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, categories)
	})

	g.GET("/category/:category/tags", func(c echo.Context) error {
		txn := newYoutubeDBTxn(database.NewTransaction(false))
		defer txn.txn.Discard()

		tags, err := txn.GetTags(c.Param("category"))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, tags)
	})
	g.GET("/category/:category/videos", func(c echo.Context) error {
		tags := strings.Split(c.QueryParam("tags"), ",")
		txn := newYoutubeDBTxn(database.NewTransaction(false))
		defer txn.txn.Discard()

		videos, err := txn.GetVideos(c.Param("category"), tags)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, videos)
	})
	adminG := g.Group("", auth.RequireMiddleware(auth.RoleAdmin))
	{
		adminG.POST("/categories", func(c echo.Context) error {
			var category YoutubeCategoryStore
			if err := c.Bind(&category); err != nil {
				return err
			}
			if category.DisplayName == "" {
				return echoerror.NewHttp(400, fmt.Errorf("display name is required"))
			}
			if category.ID == "" {
				category.ID = strings.ToLower(
					regexp.MustCompile(`[^0-9a-zA-Z]`).ReplaceAllString(category.DisplayName, "-"))
			}
			txn := newYoutubeDBTxn(database.NewTransaction(true))
			defer txn.txn.Discard()

			updatedExisting := false
			existingCategories, err := txn.GetCategories()
			if err != nil {
				if !db.IsNotFound(err) {
					return err
				}
			} else {
				for i, existingCategory := range existingCategories {
					if existingCategory.ID == category.ID {
						existingCategories[i].DisplayName = category.DisplayName
						updatedExisting = true
					}
				}
			}
			if !updatedExisting {
				existingCategories = append(existingCategories, category)
			}
			if err = txn.SetCategories(existingCategories); err != nil {
				return err
			}
			if err := txn.txn.Commit(); err != nil {
				return err
			}
			return c.JSON(http.StatusOK, category)
		})

		adminG.DELETE("/category/:category", func(c echo.Context) error {
			txn := newYoutubeDBTxn(database.NewTransaction(true))
			defer txn.txn.Discard()

			if err := txn.DeleteCategory(c.Param("category")); err != nil {
				return err
			}

			if err := txn.txn.Commit(); err != nil {
				return err
			}
			return c.NoContent(http.StatusOK)
		})

		adminG.POST("/category/:category/tags", func(c echo.Context) error {
			var tag YoutubeTagStore
			if err := c.Bind(&tag); err != nil {
				return err
			}
			if tag.DisplayName == "" {
				return echoerror.NewHttp(400, fmt.Errorf("display name is required"))
			}
			if tag.ID == "" {
				tag.ID = strings.ToLower(
					regexp.MustCompile(`[^0-9a-zA-Z]`).ReplaceAllString(tag.DisplayName, "-"))
			}
			txn := newYoutubeDBTxn(database.NewTransaction(true))
			defer txn.txn.Discard()
			updatedExisting := false

			existingTags, err := txn.GetTags(c.Param("category"))
			if err != nil {
				if !db.IsNotFound(err) {
					return err
				}
			} else {
				for i, existingTag := range existingTags {
					if existingTag.ID == tag.ID {
						existingTags[i].DisplayName = tag.DisplayName
						updatedExisting = true
					}
				}
			}
			if !updatedExisting {
				existingTags = append(existingTags, tag)
			}
			if err = txn.SetTags(c.Param("category"), existingTags); err != nil {
				return err
			}
			if err := txn.txn.Commit(); err != nil {
				return err
			}
			return c.JSON(http.StatusOK, tag)
		})

		adminG.POST("/category/:category/videos", func(c echo.Context) error {
			var video YoutubeVideoStore
			if err := c.Bind(&video); err != nil {
				return err
			}
			meta, err := GetYoutubeVideoInfo(video.VideoID)
			if err != nil {
				return err
			}
			video.Meta = meta
			txn := newYoutubeDBTxn(database.NewTransaction(true))
			defer txn.txn.Discard()
			videos, err := txn.GetVideos(c.Param("category"), nil)
			updatedExisting := false
			if err != nil {
				if !db.IsNotFound(err) {
					return err
				}
			} else {
				for i, existingVideo := range videos {
					if existingVideo.VideoID == video.VideoID {
						videos[i].Tags = video.Tags
						videos[i].Category = video.Category
						videos[i].Meta = video.Meta
						videos[i].Comment = video.Comment
						updatedExisting = true
					}
				}
			}
			if !updatedExisting {
				videos = append(videos, video)
			}
			if err := txn.SetVideos(c.Param("category"), videos); err != nil {
				return err
			}
			if err := txn.txn.Commit(); err != nil {
				return err
			}
			return c.JSON(http.StatusOK, video)
		})

		adminG.DELETE("/category/:category/video/:id", func(c echo.Context) error {
			txn := newYoutubeDBTxn(database.NewTransaction(true))
			defer txn.txn.Discard()
			videos, err := txn.GetVideos(c.Param("category"), nil)
			if err != nil {
				return err
			}
			for i, video := range videos {
				if video.VideoID == c.Param("id") {
					videos = append(videos[:i], videos[i+1:]...)
				}
			}
			if err := txn.SetVideos(c.Param("category"), videos); err != nil {
				return err
			}
			if err := txn.txn.Commit(); err != nil {
				return err
			}
			return c.NoContent(http.StatusOK)
		})
	}
}
