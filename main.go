package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

func main() {
	// Define flags
	filenamePtr := flag.String("generate-deployment", "", "Name of the output deployment file")
	appNamePtr := flag.String("app-name", "", "The name of your application")
	flag.Parse()

	if *filenamePtr != "" {
		// Check if app-name is provided
		if *appNamePtr == "" {
			fmt.Println("Please provide the application name using the -app-name flag.")
			os.Exit(1)
		}
		GenerateDeploymentFile(*filenamePtr, *appNamePtr)
		os.Exit(0)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("sample-app-%s", uuid.NewString())))
	})

	http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		allowed := q.Get("allow")
		if allowed == "true" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("preparing for shutdown"))
		} else {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("unable to shutdown"))
		}
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fmt.Println("Started server on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func GenerateDeploymentFile(filename, appName string) {
	type DeploymentConfig struct {
		AppName string
	}

	// Define the template data
	data := DeploymentConfig{AppName: appName}

	// Define the template content
	templateText := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.AppName}}-deployment
  labels:
    app: {{.AppName}}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{.AppName}}
  template:
    metadata:
      labels:
        app: {{.AppName}}
    spec:
      containers:
        - name: {{.AppName}}
          image: {{.AppName}}:latest
          imagePullPolicy: Never
          ports:
            - containerPort: 8080
              name: http-web-svc

---
apiVersion: v1
kind: Service
metadata:
  name: {{.AppName}}-svc
spec:
  type: NodePort
  selector:
    app: {{.AppName}}
  ports:
    - nodePort: 32410
      protocol: TCP
      port: 8080
      targetPort: http-web-svc
`

	// Parse the template
	tmpl, err := template.New("deployment").Parse(templateText)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	// Create the output file
	deploymentFile, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer deploymentFile.Close()

	// Execute the template and write to the file
	err = tmpl.Execute(deploymentFile, data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}
	log.Printf("%s is generate in the current working directory", filename)
}
