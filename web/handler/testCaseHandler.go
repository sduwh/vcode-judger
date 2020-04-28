package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sduwh/vcode-judger/web/util"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func ReadyUploadTestCase(c *gin.Context) {
	// 上传zip文件，预先解压至/tmp文件夹，为后续流程作准备
	caseId := c.PostForm("testCaseId")
	if caseId == "" {
		message := "The params caseId is required"
		logrus.Info(message)
		c.JSON(http.StatusBadRequest,
			NewResponse(FAIL, message, nil))
		return
	}
	logrus.Debugf("upload case for problem: %s", caseId)

	_, header, err := c.Request.FormFile("file")
	if err != nil {
		message := fmt.Sprintf("Upload file fail: %s", err)
		logrus.Info(message)
		c.JSON(http.StatusBadRequest,
			NewResponse(FAIL, message, nil))
		return
	}

	if index := strings.HasSuffix(header.Filename, ".zip"); !index {
		message := "please upload zip file"
		logrus.Info(message)
		c.JSON(http.StatusBadRequest, NewResponse(FAIL, message, nil))
		return
	}

	tmpFilePath := "/tmp"
	// save file to tmp
	if err = c.SaveUploadedFile(header, filepath.Join(tmpFilePath, header.Filename)); err != nil {
		message := fmt.Sprintf("Save zip file fail: %s", err)
		logrus.Info(message)
		c.JSON(http.StatusBadRequest, NewResponse(FAIL, message, nil))
		return
	}

	// unzip file to tmp
	testCaseId, err := util.Unzip(header.Filename, tmpFilePath, tmpFilePath+"/"+caseId)
	if err != nil {
		message := fmt.Sprintf("unzip file fail: %s", err)
		logrus.Info(message)
		c.JSON(http.StatusBadRequest,
			NewResponse(FAIL, message, nil))
		return
	}

	logrus.Info("Upload file success")
	c.JSON(http.StatusOK, NewResponse(SUCCESS, "Upload file success", testCaseId))
}

func CheckCase(c *gin.Context) {
	// 从tmp文件夹中将测试用例移动至目标data文件夹
	caseId := c.PostForm("testCaseId")
	if caseId == "" {
		message := "The params caseId is required"
		logrus.Info(message)
		c.JSON(http.StatusBadRequest,
			NewResponse(FAIL, message, nil))
		return
	}
	oldCaseId := c.PostForm("oldTestCaseId")

	logrus.Debugf("check case for problem: %s", caseId)

	casePath := filepath.Join("/tmp", caseId)

	for {
		if err := os.Rename(casePath, filepath.Join("./cases", caseId)); err != nil {
			if strings.Contains(err.Error(), "file exists") {
				_ = os.RemoveAll(filepath.Join("./cases", caseId))
				continue
			}
			message := fmt.Sprintf("Move case dir fail: %s", err)
			logrus.Error(message)
			c.JSON(http.StatusBadRequest, NewResponse(FAIL, message, nil))
			return
		}
		break
	}
	if oldCaseId == "" {
		for {
			err := os.RemoveAll(filepath.Join("./cases", oldCaseId))
			if err == nil {
				break
			}
		}
	}

	c.JSON(http.StatusOK, NewResponse(SUCCESS, "success", nil))
}

func TestHandlerRoutes(router *gin.Engine) {
	routes := router.Group("/api/case")
	{
		routes.POST("/ready", ReadyUploadTestCase)
		routes.POST("/check", CheckCase)
	}
}
