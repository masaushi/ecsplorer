package ai

import "fmt"

// BuildLogAnalysisPrompt builds a prompt for analyzing ECS service logs.
func BuildLogAnalysisPrompt(serviceName, clusterName, logEntries, taskDefSummary string) string {
	return fmt.Sprintf(`You are an AWS ECS expert. Analyze the following CloudWatch logs for the ECS service.

Service: %s
Cluster: %s

Task Definition Summary:
%s

Recent Log Entries:
%s

Please provide your analysis using the following format with tview color tags for formatting:

[yellow]== Log Analysis Report ==[white]

[blue]1. Overview[white]
Summarize the overall state of the service based on the logs.

[blue]2. Errors & Warnings[white]
List any errors or warnings found. Use [red] for errors and [yellow] for warnings.

[blue]3. Patterns[white]
Identify any recurring patterns, frequency of errors, or anomalies.

[blue]4. Recommendations[white]
Provide actionable recommendations to address any issues found.
Use [green] to highlight positive findings and best practices.`, serviceName, clusterName, taskDefSummary, logEntries)
}

// BuildMetricsAnalysisPrompt builds a prompt for analyzing ECS service metrics.
func BuildMetricsAnalysisPrompt(serviceName, clusterName, metricsData, serviceConfig string) string {
	return fmt.Sprintf(`You are an AWS ECS expert. Analyze the following CloudWatch metrics for the ECS service.

Service: %s
Cluster: %s

Service Configuration:
%s

Metrics Data:
%s

Please provide your analysis using the following format with tview color tags for formatting:

[yellow]== Metrics Analysis Report ==[white]

[blue]1. Resource Utilization Summary[white]
Summarize CPU and memory utilization trends.
Use [red] for critical levels (>90%%), [yellow] for warning levels (>70%%), and [green] for healthy levels.

[blue]2. Trends & Patterns[white]
Identify usage trends, spikes, or unusual patterns over the time period.

[blue]3. Scaling Assessment[white]
Evaluate if the current task count and resource allocation are appropriate.

[blue]4. Cost Optimization[white]
Suggest any opportunities to optimize resource allocation and reduce costs.

[blue]5. Recommendations[white]
Provide specific, actionable recommendations for improving performance or efficiency.`, serviceName, clusterName, serviceConfig, metricsData)
}

// BuildConfigReviewPrompt builds a prompt for reviewing ECS configuration.
func BuildConfigReviewPrompt(entityName, entityType, configData string) string {
	return fmt.Sprintf(`You are an AWS ECS expert. Review the following ECS %s configuration for best practices and potential issues.

%s Name: %s

Configuration:
%s

Please provide your review using the following format with tview color tags for formatting:

[yellow]== Configuration Review Report ==[white]

[blue]1. Configuration Summary[white]
Brief overview of the current configuration.

[blue]2. Security Review[white]
Check for security best practices. Use [red] for security concerns and [green] for good practices.

[blue]3. Reliability Review[white]
Assess high availability, health checks, and fault tolerance settings.

[blue]4. Performance Review[white]
Evaluate resource allocation, scaling configuration, and performance settings.

[blue]5. Best Practices[white]
Check alignment with AWS ECS best practices. Use [yellow] for suggestions.

[blue]6. Recommendations[white]
Prioritized list of recommended changes. Use [red] for critical, [yellow] for important, and [green] for nice-to-have.`, entityType, entityType, entityName, configData)
}

// BuildTroubleshootPrompt builds a prompt for troubleshooting ECS service issues.
func BuildTroubleshootPrompt(serviceName, clusterName, diagnosticData string) string {
	return fmt.Sprintf(`You are an AWS ECS expert troubleshooter. Analyze the following diagnostic data for the ECS service and help identify and resolve any issues.

Service: %s
Cluster: %s

Diagnostic Data:
%s

Please provide your troubleshooting analysis using the following format with tview color tags for formatting:

[yellow]== Troubleshooting Report ==[white]

[blue]1. Health Assessment[white]
Overall health status of the service.
Use [red] for critical issues, [yellow] for warnings, and [green] for healthy components.

[blue]2. Issues Found[white]
List all identified issues, ordered by severity.
Use [red]CRITICAL[white], [yellow]WARNING[white], or [blue]INFO[white] tags for each issue.

[blue]3. Root Cause Analysis[white]
For each issue, provide a likely root cause explanation.

[blue]4. Service Events Analysis[white]
Analyze recent service events for deployment issues or failures.

[blue]5. Resolution Steps[white]
Step-by-step instructions to resolve each identified issue.
Number each step and use [green] for completed/verified steps.

[blue]6. Preventive Measures[white]
Suggestions to prevent similar issues in the future.`, serviceName, clusterName, diagnosticData)
}
