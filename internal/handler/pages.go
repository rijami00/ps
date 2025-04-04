package handler

import (
	"github.com/haatos/goshipit/internal/apollo"
	"log"
	"net/http"
	"os"

	"github.com/haatos/goshipit/internal"
	"github.com/haatos/goshipit/internal/markdown"
	"github.com/haatos/goshipit/internal/views/pages"
	"github.com/labstack/echo/v4"
)

var gettingStartedHTML string
var typesHTML string

const (
	contentTypesMarkdownPath   = "content/types.md"
	gettingStartedMarkdownPath = "content/getting_started.md"
)

func getGettingStartedHTML() {
	if gettingStartedHTML != "" {
		return
	}

	pageContent, err := os.ReadFile(gettingStartedMarkdownPath)
	if err != nil {
		log.Fatal(err)
	}

	gettingStartedHTML = markdown.GetHTMLFromMarkdown(pageContent)
}

func getTypesHTML() {
	if typesHTML != "" {
		return
	}

	pageContent, err := os.ReadFile(contentTypesMarkdownPath)
	if err != nil {
		log.Fatal(err)
	}

	typesHTML = markdown.GetHTMLFromMarkdown(pageContent)
}

func GetIndexPage(c echo.Context) error {
	// var content, _ = apollo.PrintInstancesHTML()
	var instances, _ = apollo.GetInstances()
	if isHXRequest(c) {
		return render(c, http.StatusOK, pages.IndexPageContent(instances))
	}
	return render(c, http.StatusOK, pages.IndexPage(instances))
}

func GetJsonApi(c echo.Context) error {
	var instances, _ = apollo.GetInstances()
	return c.JSON(http.StatusOK, instances)
}

func GetAboutPage(c echo.Context) error {
	if isHXRequest(c) {
		return render(c, http.StatusOK, pages.AboutPageMain())
	}
	return render(c, http.StatusOK, pages.AboutPage())
}

func GetGettingStartedPage(c echo.Context) error {
	getGettingStartedHTML()

	if isHXRequest(c) {
		return render(c, http.StatusOK, pages.GettingStartedPageMain(gettingStartedHTML))
	}
	return render(c, http.StatusOK, pages.GettingStartedPage(gettingStartedHTML))
}

func GetTypesPage(c echo.Context) error {
	getTypesHTML()

	if isHXRequest(c) {
		return render(c, http.StatusOK, pages.TypesPageMain(typesHTML))
	}
	return render(c, http.StatusOK, pages.TypesPage(typesHTML))
}

func GetPrivacyPolicyPage(c echo.Context) error {
	if isHXRequest(c) {
		return render(
			c, http.StatusOK,
			pages.PrivacyMain(internal.Settings.Domain, internal.Settings.ContactEmail))
	}
	return render(
		c, http.StatusOK,
		pages.PrivacyPage(internal.Settings.Domain, internal.Settings.ContactEmail))
}

func GetTermsOfServicePage(c echo.Context) error {
	if isHXRequest(c) {
		return render(
			c, http.StatusOK,
			pages.TermsOfServiceMain(internal.Settings.Domain, internal.Settings.ContactEmail))
	}
	return render(
		c, http.StatusOK,
		pages.TermsOfService(internal.Settings.Domain, internal.Settings.ContactEmail))
}
