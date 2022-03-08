package mcaf

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAWSCodeBuildTrigger() *schema.Resource {
	return &schema.Resource{
		Create: checkProvider("aws", resourceAWSCodeBuildTriggerCreate),
		Read:   checkProvider("aws", resourceAWSCodeBuildTriggerRead),
		Update: checkProvider("aws", resourceAWSCodeBuildTriggerUpdate),
		Delete: checkProvider("aws", resourceAWSCodeBuildTriggerDelete),

		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"release_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAWSCodeBuildTriggerCreate(d *schema.ResourceData, meta interface{}) error {
	// Set the ID.
	d.SetId(d.Get("project").(string))

	// Trigger the CodeBuild pipeline.
	return triggerCodeBuildPipeline(d, meta)
}

func resourceAWSCodeBuildTriggerRead(d *schema.ResourceData, meta interface{}) error {
	// There isn't anything to read back.
	return nil
}

func resourceAWSCodeBuildTriggerUpdate(d *schema.ResourceData, meta interface{}) error {
	// Trigger the CodeBuild pipeline.
	return triggerCodeBuildPipeline(d, meta)
}

func resourceAWSCodeBuildTriggerDelete(d *schema.ResourceData, meta interface{}) error {
	// There isn't anything to delete.
	return nil
}

func triggerCodeBuildPipeline(d *schema.ResourceData, meta interface{}) error {
	cbconn := meta.(*Client).AWSClient.cbconn

	// Define the required input parameters.
	input := &codebuild.StartBuildInput{
		ProjectName:   aws.String(d.Get("project").(string)),
		SourceVersion: aws.String(d.Get("version").(string)),
	}

	// Trigger all pipelines by starting a new build.
	if _, err := cbconn.StartBuild(input); err != nil {
		return fmt.Errorf("Failed to start new build: %v", err)
	}

	return nil
}
