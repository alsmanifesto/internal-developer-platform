// Package generator handles file scaffolding for new services.
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ravon/scaffold/internal/metadata"
)

// Generate creates the full scaffolding for a service.
func Generate(svc metadata.ServiceMetadata) error {
	base := svc.Name

	if err := os.MkdirAll(base, 0o755); err != nil {
		return fmt.Errorf("create service dir: %w", err)
	}

	if err := writeStackFiles(base, svc); err != nil {
		return fmt.Errorf("write stack files: %w", err)
	}

	if err := writeReadme(base, svc); err != nil {
		return fmt.Errorf("write README: %w", err)
	}

	if err := writeDockerfile(base, svc); err != nil {
		return fmt.Errorf("write Dockerfile: %w", err)
	}

	if err := writeTerraform(base, svc); err != nil {
		return fmt.Errorf("write Terraform: %w", err)
	}

	if err := writePipeline(base, svc); err != nil {
		return fmt.Errorf("write pipeline: %w", err)
	}

	return nil
}

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func writeStackFiles(base string, svc metadata.ServiceMetadata) error {
	switch svc.Stack {
	case "go":
		content := "package main\n\nimport (\n\t\"fmt\"\n\t\"log\"\n\t\"net/http\"\n)\n\nfunc main() {\n\thttp.HandleFunc(\"/\", func(w http.ResponseWriter, r *http.Request) {\n\t\tfmt.Fprintln(w, \"Hello, World!\")\n\t})\n\n\tlog.Println(\"Server starting on :8080\")\n\tif err := http.ListenAndServe(\":8080\", nil); err != nil {\n\t\tlog.Fatal(err)\n\t}\n}\n"
		return writeFile(filepath.Join(base, "cmd", "main.go"), content)

	case "python":
		content := "#!/usr/bin/env python3\n\"\"\"Simple hello world service.\"\"\"\n\n\ndef main():\n    print(\"Hello, World!\")\n\n\nif __name__ == \"__main__\":\n    main()\n"
		return writeFile(filepath.Join(base, "main.py"), content)

	case "spark":
		content := "#!/usr/bin/env python3\n\"\"\"Basic PySpark job with a simple transformation.\"\"\"\n\nfrom pyspark.sql import SparkSession\nfrom pyspark.sql.functions import col, upper\n\n\ndef main():\n    spark = SparkSession.builder \\\n        .appName(\"scaffold-spark-job\") \\\n        .getOrCreate()\n\n    data = [(\"hello\", 1), (\"world\", 2), (\"spark\", 3)]\n    df = spark.createDataFrame(data, [\"word\", \"count\"])\n\n    result = df.withColumn(\"upper_word\", upper(col(\"word\")))\n    result.show()\n\n    spark.stop()\n\n\nif __name__ == \"__main__\":\n    main()\n"
		return writeFile(filepath.Join(base, "main.py"), content)

	case "kafka":
		producer := "#!/usr/bin/env python3\n\"\"\"Simple Kafka producer example.\"\"\"\n\nimport json\nimport time\nfrom kafka import KafkaProducer\n\n\ndef main():\n    producer = KafkaProducer(\n        bootstrap_servers=[\"localhost:9092\"],\n        value_serializer=lambda v: json.dumps(v).encode(\"utf-8\"),\n    )\n\n    topic = \"scaffold-topic\"\n    for i in range(10):\n        message = {\"index\": i, \"message\": f\"Hello from Scaffold producer #{i}\"}\n        producer.send(topic, message)\n        print(f\"Sent: {message}\")\n        time.sleep(1)\n\n    producer.flush()\n    producer.close()\n\n\nif __name__ == \"__main__\":\n    main()\n"
		consumer := "#!/usr/bin/env python3\n\"\"\"Simple Kafka consumer example.\"\"\"\n\nfrom kafka import KafkaConsumer\nimport json\n\n\ndef main():\n    consumer = KafkaConsumer(\n        \"scaffold-topic\",\n        bootstrap_servers=[\"localhost:9092\"],\n        auto_offset_reset=\"earliest\",\n        enable_auto_commit=True,\n        group_id=\"scaffold-consumer-group\",\n        value_deserializer=lambda m: json.loads(m.decode(\"utf-8\")),\n    )\n\n    print(\"Waiting for messages...\")\n    for message in consumer:\n        print(f\"Received: {message.value}\")\n\n\nif __name__ == \"__main__\":\n    main()\n"
		if err := writeFile(filepath.Join(base, "producer.py"), producer); err != nil {
			return err
		}
		return writeFile(filepath.Join(base, "consumer.py"), consumer)
	}

	return nil
}

func writeReadme(base string, svc metadata.ServiceMetadata) error {
	runInstructions := runInstructionsFor(svc.Stack)
	dockerInstructions := dockerInstructionsFor(svc.Name)

	var lines []string
	lines = append(lines, "# "+svc.Name)
	lines = append(lines, "")
	lines = append(lines, "> Scaffolded by [Scaffold](https://github.com/ravon/scaffold) on "+time.Now().Format("2006-01-02"))
	lines = append(lines, "")
	lines = append(lines, "## Service Configuration")
	lines = append(lines, "")
	lines = append(lines, "| Field        | Value         |")
	lines = append(lines, "|--------------|---------------|")
	lines = append(lines, fmt.Sprintf("| Name         | %-13s |", svc.Name))
	lines = append(lines, fmt.Sprintf("| Service Type | %-13s |", svc.ServiceType))
	lines = append(lines, fmt.Sprintf("| Workload     | %-13s |", svc.Workload))
	lines = append(lines, fmt.Sprintf("| Stack        | %-13s |", svc.Stack))
	lines = append(lines, fmt.Sprintf("| Pipeline     | %-13s |", svc.Pipeline))
	lines = append(lines, "")
	lines = append(lines, "## Description")
	lines = append(lines, "")
	lines = append(lines, "This service was bootstrapped using Scaffold, the internal developer platform CLI.")
	lines = append(lines, fmt.Sprintf("It is a **%s** workload running as a **%s** using **%s**.", svc.Workload, svc.ServiceType, svc.Stack))
	lines = append(lines, "")
	lines = append(lines, "## Running the Service")
	lines = append(lines, "")
	lines = append(lines, runInstructions)
	lines = append(lines, "")
	lines = append(lines, "## Docker")
	lines = append(lines, "")
	lines = append(lines, dockerInstructions)
	lines = append(lines, "")
	lines = append(lines, "## Infrastructure")
	lines = append(lines, "")
	lines = append(lines, "Terraform configuration is available in the `terraform/` directory.")
	lines = append(lines, "")
	lines = append(lines, "```bash")
	lines = append(lines, "cd terraform")
	lines = append(lines, "terraform init")
	lines = append(lines, "terraform apply")
	lines = append(lines, "```")
	lines = append(lines, "")
	lines = append(lines, "> Note: Configured for LocalStack compatibility.")
	lines = append(lines, "")

	content := strings.Join(lines, "\n")
	return writeFile(filepath.Join(base, "README.md"), content)
}

func runInstructionsFor(stack string) string {
	switch stack {
	case "go":
		return "```bash\ngo run cmd/main.go\n```"
	case "python":
		return "```bash\npython3 main.py\n```"
	case "spark":
		return "```bash\n# Ensure PySpark is installed\npip install pyspark\npython3 main.py\n```"
	case "kafka":
		return "```bash\n# Ensure kafka-python is installed\npip install kafka-python\n\n# Start producer\npython3 producer.py\n\n# Start consumer (in another terminal)\npython3 consumer.py\n```"
	}
	return ""
}

func dockerInstructionsFor(name string) string {
	return fmt.Sprintf("```bash\ndocker build -t %s .\ndocker run --rm %s\n```", name, name)
}

func writeDockerfile(base string, svc metadata.ServiceMetadata) error {
	var lines []string

	switch svc.Stack {
	case "go":
		lines = []string{
			"# Multi-stage build for Go service",
			"FROM golang:1.21-alpine AS builder",
			"",
			"WORKDIR /app",
			"COPY . .",
			"RUN go mod download",
			"RUN go build -o /app/server ./cmd/main.go",
			"",
			"FROM alpine:latest",
			"",
			"RUN apk --no-cache add ca-certificates",
			"WORKDIR /root/",
			"",
			"COPY --from=builder /app/server .",
			"",
			"EXPOSE 8080",
			`CMD ["./server"]`,
			"",
		}
	case "python":
		lines = []string{
			"FROM python:3.11-slim",
			"",
			"WORKDIR /app",
			"COPY . .",
			"",
			`CMD ["python3", "main.py"]`,
			"",
		}
	case "spark":
		lines = []string{
			"FROM python:3.11-slim",
			"",
			"RUN pip install --no-cache-dir pyspark",
			"",
			"WORKDIR /app",
			"COPY . .",
			"",
			`CMD ["python3", "main.py"]`,
			"",
		}
	case "kafka":
		lines = []string{
			"FROM python:3.11-slim",
			"",
			"RUN pip install --no-cache-dir kafka-python",
			"",
			"WORKDIR /app",
			"COPY . .",
			"",
			`CMD ["python3", "producer.py"]`,
			"",
		}
	}

	content := strings.Join(lines, "\n")
	return writeFile(filepath.Join(base, "Dockerfile"), content)
}

func writeTerraform(base string, svc metadata.ServiceMetadata) error {
	bucketName := strings.ToLower(strings.ReplaceAll(svc.Name, "_", "-"))

	var lines []string
	lines = append(lines, "terraform {")
	lines = append(lines, `  required_version = ">= 1.0"`)
	lines = append(lines, "")
	lines = append(lines, "  required_providers {")
	lines = append(lines, "    aws = {")
	lines = append(lines, `      source  = "hashicorp/aws"`)
	lines = append(lines, `      version = "~> 5.0"`)
	lines = append(lines, "    }")
	lines = append(lines, "  }")
	lines = append(lines, "}")
	lines = append(lines, "")
	lines = append(lines, "# LocalStack-compatible AWS provider configuration")
	lines = append(lines, `provider "aws" {`)
	lines = append(lines, `  region                      = "us-east-1"`)
	lines = append(lines, `  access_key                  = "test"`)
	lines = append(lines, `  secret_key                  = "test"`)
	lines = append(lines, "  skip_credentials_validation = true")
	lines = append(lines, "  skip_metadata_api_check     = true")
	lines = append(lines, "  skip_requesting_account_id  = true")
	lines = append(lines, "")
	lines = append(lines, "  s3_use_path_style = true")
	lines = append(lines, "")
	lines = append(lines, "  endpoints {")
	lines = append(lines, `    s3 = "http://localhost:4566"`)
	lines = append(lines, "  }")
	lines = append(lines, "}")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf(`resource "aws_s3_bucket" "%s" {`, bucketName))
	lines = append(lines, fmt.Sprintf(`  bucket = "%s-bucket"`, bucketName))
	lines = append(lines, "")
	lines = append(lines, "  tags = {")
	lines = append(lines, fmt.Sprintf(`    Name        = "%s"`, svc.Name))
	lines = append(lines, fmt.Sprintf(`    ServiceType = "%s"`, svc.ServiceType))
	lines = append(lines, fmt.Sprintf(`    Workload    = "%s"`, svc.Workload))
	lines = append(lines, fmt.Sprintf(`    Stack       = "%s"`, svc.Stack))
	lines = append(lines, `    ManagedBy   = "scaffold"`)
	lines = append(lines, "  }")
	lines = append(lines, "}")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf(`resource "aws_s3_bucket_versioning" "%s_versioning" {`, bucketName))
	lines = append(lines, fmt.Sprintf("  bucket = aws_s3_bucket.%s.id", bucketName))
	lines = append(lines, "")
	lines = append(lines, "  versioning_configuration {")
	lines = append(lines, `    status = "Enabled"`)
	lines = append(lines, "  }")
	lines = append(lines, "}")
	lines = append(lines, "")
	lines = append(lines, "output \"bucket_name\" {")
	lines = append(lines, fmt.Sprintf("  value = aws_s3_bucket.%s.bucket", bucketName))
	lines = append(lines, "}")
	lines = append(lines, "")

	content := strings.Join(lines, "\n")
	return writeFile(filepath.Join(base, "terraform", "main.tf"), content)
}

func writePipeline(base string, svc metadata.ServiceMetadata) error {
	switch svc.Pipeline {
	case "gh-actions":
		return writeGHActions(base, svc)
	case "concourse":
		return writeConcourse(base, svc)
	case "airflow":
		return writeAirflow(base, svc)
	case "mlflow":
		return writeMLflow(base, svc)
	}
	return nil
}

func writeGHActions(base string, svc metadata.ServiceMetadata) error {
	var buildStepLines []string
	switch svc.Stack {
	case "go":
		buildStepLines = []string{
			"      - name: Set up Go",
			"        uses: actions/setup-go@v4",
			"        with:",
			"          go-version: '1.21'",
			"",
			"      - name: Build",
			"        run: go build ./...",
			"",
			"      - name: Test",
			"        run: go test ./...",
		}
	default:
		buildStepLines = []string{
			"      - name: Set up Python",
			"        uses: actions/setup-python@v4",
			"        with:",
			"          python-version: '3.11'",
			"",
			"      - name: Install dependencies",
			"        run: pip install -r requirements.txt || true",
			"",
			"      - name: Lint",
			"        run: pip install flake8 && flake8 . --count --select=E9,F63,F7,F82 --show-source --statistics || true",
			"",
			"      - name: Test",
			"        run: python -m pytest || true",
		}
	}

	var lines []string
	lines = append(lines, "name: CI")
	lines = append(lines, "")
	lines = append(lines, "on:")
	lines = append(lines, "  push:")
	lines = append(lines, "    branches: [ main ]")
	lines = append(lines, "  pull_request:")
	lines = append(lines, "    branches: [ main ]")
	lines = append(lines, "")
	lines = append(lines, "jobs:")
	lines = append(lines, "  build-and-test:")
	lines = append(lines, "    runs-on: ubuntu-latest")
	lines = append(lines, "")
	lines = append(lines, "    steps:")
	lines = append(lines, "      - name: Checkout code")
	lines = append(lines, "        uses: actions/checkout@v4")
	lines = append(lines, "")
	lines = append(lines, buildStepLines...)
	lines = append(lines, "")
	lines = append(lines, "      - name: Build Docker image")
	lines = append(lines, fmt.Sprintf("        run: docker build -t %s .", svc.Name))
	lines = append(lines, "")

	content := strings.Join(lines, "\n")
	return writeFile(filepath.Join(base, ".github", "workflows", "ci.yml"), content)
}

func writeConcourse(base string, svc metadata.ServiceMetadata) error {
	var lines []string
	lines = append(lines, "---")
	lines = append(lines, "resources:")
	lines = append(lines, "  - name: source-code")
	lines = append(lines, "    type: git")
	lines = append(lines, "    source:")
	lines = append(lines, fmt.Sprintf("      uri: https://github.com/your-org/%s", svc.Name))
	lines = append(lines, "      branch: main")
	lines = append(lines, "")
	lines = append(lines, "jobs:")
	lines = append(lines, "  - name: build-and-test")
	lines = append(lines, "    plan:")
	lines = append(lines, "      - get: source-code")
	lines = append(lines, "        trigger: true")
	lines = append(lines, "")
	lines = append(lines, "      - task: build")
	lines = append(lines, "        config:")
	lines = append(lines, "          platform: linux")
	lines = append(lines, "          image_resource:")
	lines = append(lines, "            type: registry-image")
	lines = append(lines, "            source:")
	lines = append(lines, "              repository: alpine")
	lines = append(lines, "          inputs:")
	lines = append(lines, "            - name: source-code")
	lines = append(lines, "          run:")
	lines = append(lines, "            path: sh")
	lines = append(lines, "            args:")
	lines = append(lines, "              - -c")
	lines = append(lines, "              - |")
	lines = append(lines, fmt.Sprintf(`                echo "Building %s..."`, svc.Name))
	lines = append(lines, fmt.Sprintf(`                echo "Stack: %s"`, svc.Stack))
	lines = append(lines, `                echo "Build complete."`)
	lines = append(lines, "")
	lines = append(lines, "  - name: deploy")
	lines = append(lines, "    plan:")
	lines = append(lines, "      - get: source-code")
	lines = append(lines, "        trigger: false")
	lines = append(lines, "        passed: [build-and-test]")
	lines = append(lines, "")
	lines = append(lines, "      - task: deploy")
	lines = append(lines, "        config:")
	lines = append(lines, "          platform: linux")
	lines = append(lines, "          image_resource:")
	lines = append(lines, "            type: registry-image")
	lines = append(lines, "            source:")
	lines = append(lines, "              repository: alpine")
	lines = append(lines, "          inputs:")
	lines = append(lines, "            - name: source-code")
	lines = append(lines, "          run:")
	lines = append(lines, "            path: sh")
	lines = append(lines, "            args:")
	lines = append(lines, "              - -c")
	lines = append(lines, fmt.Sprintf(`              - echo "Deploying %s..."`, svc.Name))
	lines = append(lines, "")

	content := strings.Join(lines, "\n")
	return writeFile(filepath.Join(base, "pipeline.yml"), content)
}

func writeAirflow(base string, svc metadata.ServiceMetadata) error {
	safeName := strings.ReplaceAll(svc.Name, "-", "_")

	var lines []string
	lines = append(lines, `"""`)
	lines = append(lines, fmt.Sprintf("Airflow DAG for %s service.", svc.Name))
	lines = append(lines, "Generated by Scaffold.")
	lines = append(lines, `"""`)
	lines = append(lines, "")
	lines = append(lines, "from datetime import datetime, timedelta")
	lines = append(lines, "")
	lines = append(lines, "from airflow import DAG")
	lines = append(lines, "from airflow.operators.python import PythonOperator")
	lines = append(lines, "from airflow.operators.bash import BashOperator")
	lines = append(lines, "")
	lines = append(lines, "")
	lines = append(lines, "default_args = {")
	lines = append(lines, `    "owner": "scaffold",`)
	lines = append(lines, `    "depends_on_past": False,`)
	lines = append(lines, `    "email_on_failure": False,`)
	lines = append(lines, `    "email_on_retry": False,`)
	lines = append(lines, `    "retries": 1,`)
	lines = append(lines, `    "retry_delay": timedelta(minutes=5),`)
	lines = append(lines, "}")
	lines = append(lines, "")
	lines = append(lines, "with DAG(")
	lines = append(lines, fmt.Sprintf(`    dag_id="%s_pipeline",`, safeName))
	lines = append(lines, "    default_args=default_args,")
	lines = append(lines, fmt.Sprintf(`    description="Pipeline for %s",`, svc.Name))
	lines = append(lines, `    schedule_interval="@daily",`)
	lines = append(lines, "    start_date=datetime(2024, 1, 1),")
	lines = append(lines, "    catchup=False,")
	lines = append(lines, fmt.Sprintf(`    tags=["%s", "%s", "%s"],`, svc.Stack, svc.Workload, svc.ServiceType))
	lines = append(lines, ") as dag:")
	lines = append(lines, "")
	lines = append(lines, "    def extract(**kwargs):")
	lines = append(lines, fmt.Sprintf(`        print("Extracting data for %s...")`, svc.Name))
	lines = append(lines, `        return {"status": "extracted"}`)
	lines = append(lines, "")
	lines = append(lines, "    def transform(**kwargs):")
	lines = append(lines, fmt.Sprintf(`        print("Transforming data for %s...")`, svc.Name))
	lines = append(lines, `        return {"status": "transformed"}`)
	lines = append(lines, "")
	lines = append(lines, "    def load(**kwargs):")
	lines = append(lines, fmt.Sprintf(`        print("Loading data for %s...")`, svc.Name))
	lines = append(lines, `        return {"status": "loaded"}`)
	lines = append(lines, "")
	lines = append(lines, "    extract_task = PythonOperator(")
	lines = append(lines, `        task_id="extract",`)
	lines = append(lines, "        python_callable=extract,")
	lines = append(lines, "    )")
	lines = append(lines, "")
	lines = append(lines, "    transform_task = PythonOperator(")
	lines = append(lines, `        task_id="transform",`)
	lines = append(lines, "        python_callable=transform,")
	lines = append(lines, "    )")
	lines = append(lines, "")
	lines = append(lines, "    load_task = PythonOperator(")
	lines = append(lines, `        task_id="load",`)
	lines = append(lines, "        python_callable=load,")
	lines = append(lines, "    )")
	lines = append(lines, "")
	lines = append(lines, "    done_task = BashOperator(")
	lines = append(lines, `        task_id="done",`)
	lines = append(lines, fmt.Sprintf(`        bash_command='echo "Pipeline for %s completed successfully."',`, svc.Name))
	lines = append(lines, "    )")
	lines = append(lines, "")
	lines = append(lines, "    extract_task >> transform_task >> load_task >> done_task")
	lines = append(lines, "")

	content := strings.Join(lines, "\n")
	return writeFile(filepath.Join(base, "dags", "dag.py"), content)
}

func writeMLflow(base string, svc metadata.ServiceMetadata) error {
	var lines []string
	lines = append(lines, `"""`)
	lines = append(lines, fmt.Sprintf("MLflow pipeline placeholder for %s.", svc.Name))
	lines = append(lines, "Generated by Scaffold.")
	lines = append(lines, `"""`)
	lines = append(lines, "")
	lines = append(lines, "import mlflow")
	lines = append(lines, "import mlflow.sklearn")
	lines = append(lines, "from datetime import datetime")
	lines = append(lines, "")
	lines = append(lines, "")
	lines = append(lines, "def run_pipeline():")
	lines = append(lines, `    """Main ML pipeline entry point."""`)
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf(`    experiment_name = "%s-experiment"`, svc.Name))
	lines = append(lines, "    mlflow.set_experiment(experiment_name)")
	lines = append(lines, "")
	lines = append(lines, `    with mlflow.start_run(run_name=f"run-{datetime.now().strftime('%Y%m%d-%H%M%S')}"):`)
	lines = append(lines, "        # Log pipeline parameters")
	lines = append(lines, fmt.Sprintf(`        mlflow.log_param("service_name", "%s")`, svc.Name))
	lines = append(lines, fmt.Sprintf(`        mlflow.log_param("workload", "%s")`, svc.Workload))
	lines = append(lines, fmt.Sprintf(`        mlflow.log_param("stack", "%s")`, svc.Stack))
	lines = append(lines, "")
	lines = append(lines, "        # --- Data ingestion ---")
	lines = append(lines, `        print("Step 1: Loading data...")`)
	lines = append(lines, "        # TODO: Replace with your data loading logic")
	lines = append(lines, "")
	lines = append(lines, "        # --- Feature engineering ---")
	lines = append(lines, `        print("Step 2: Feature engineering...")`)
	lines = append(lines, "        # TODO: Add your feature transformations here")
	lines = append(lines, "")
	lines = append(lines, "        # --- Model training ---")
	lines = append(lines, `        print("Step 3: Training model...")`)
	lines = append(lines, "        # TODO: Replace with your model training logic")
	lines = append(lines, "")
	lines = append(lines, "        # Log placeholder metrics")
	lines = append(lines, "        mlflow.log_metric(\"accuracy\", 0.0)")
	lines = append(lines, "        mlflow.log_metric(\"loss\", 0.0)")
	lines = append(lines, "")
	lines = append(lines, "        # --- Model evaluation ---")
	lines = append(lines, `        print("Step 4: Evaluating model...")`)
	lines = append(lines, "        # TODO: Add evaluation logic here")
	lines = append(lines, "")
	lines = append(lines, "        # --- Model registration ---")
	lines = append(lines, `        print("Step 5: Registering model...")`)
	lines = append(lines, fmt.Sprintf("        # TODO: mlflow.sklearn.log_model(model, \"%s-model\")", svc.Name))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf(`        print(f"Pipeline for %s completed. Check MLflow UI for results.")`, svc.Name))
	lines = append(lines, "")
	lines = append(lines, "")
	lines = append(lines, `if __name__ == "__main__":`)
	lines = append(lines, "    run_pipeline()")
	lines = append(lines, "")

	content := strings.Join(lines, "\n")
	return writeFile(filepath.Join(base, "mlflow_pipeline.py"), content)
}
