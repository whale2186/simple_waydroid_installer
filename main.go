package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	if len(os.Args) != 2 {		
		log.Fatalf("Usage: %s <APK FILE>",os.Args[0])
		os.Exit(1)
	}

	apkPath := os.Args[1]  // Replace with the path to your APK file

	appName, packageName, versionName, versionCode, icon, err := parseAPKWithZip(apkPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	a := app.New()
	windowname := "APK Installer"
	w := a.NewWindow(windowname)
	appsrc := widget.NewLabel(apkPath)
	appNameLabel := widget.NewLabel(appName)
	packageNameLabel := widget.NewLabel(packageName)
	versionNameLabel := widget.NewLabel(versionName)
	versionCodeLabel := widget.NewLabel(versionCode)
	iconshow := canvas.NewImageFromFile(icon)
	w.SetContent(container.NewVBox(
		iconshow,
		appsrc,
		appNameLabel,
		packageNameLabel,
		versionNameLabel,
		versionCodeLabel,
		widget.NewButton("Install APK", func() {
			installapk(apkPath)	
		}),
	))
	w.ShowAndRun()

}

type APKManifest struct {
	XMLName      xml.Name    `xml:"manifest"`
	Package      string      `xml:"package,attr"`
	VersionName  string      `xml:"versionName,attr"`
	VersionCode  string      `xml:"versionCode,attr"`
	Application  Application `xml:"application"`
}

type Application struct {
	Icon  string `xml:"icon,attr"`
	Label string `xml:"label,attr"`
}

func parseAPKWithZip(apkPath string) (appName, packageName, versionName, versionCode, icon string, err error) {
	r, err := zip.OpenReader(apkPath)
	if err != nil {
		return "", "", "", "", "", err
	}
	defer r.Close()

	var manifestData, stringsData []byte

	for _, f := range r.File {
		if f.Name == "AndroidManifest.xml" {
			manifestData, err = readZipFile(f)
			if err != nil {
				return "", "", "", "", "", err
			}
		} else if f.Name == "res/values/strings.xml" {
			stringsData, err = readZipFile(f)
			if err != nil {
				return "", "", "", "", "", err
			}
		}
	}

	if manifestData == nil {
		return "", "", "", "", "", fmt.Errorf("AndroidManifest.xml not found")
	}

	var manifest APKManifest
	err = xml.Unmarshal(manifestData, &manifest)
	if err != nil {
		return "", "", "", "", "", err
	}

	appName = extractAppName(stringsData, manifest.Application.Label)
	packageName = manifest.Package
	versionName = manifest.VersionName
	versionCode = manifest.VersionCode
	icon = manifest.Application.Icon

	return appName, packageName, versionName, versionCode, icon, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return ioutil.ReadAll(rc)
}

func extractAppName(stringsData []byte, label string) string {
	if strings.HasPrefix(label, "@string/") {
		re := regexp.MustCompile(fmt.Sprintf(`<string name="%s">([^<]*)</string>`, strings.TrimPrefix(label, "@string/")))
		matches := re.FindSubmatch(stringsData)
		if len(matches) > 1 {
			return string(matches[1])
		}
	}
	return label
}

func installapk(applocation string)(msg string)  {
	exec.Command("waydroid", "app", "install", applocation)
	msg = "App installed sucessfully !"
	return msg 
}