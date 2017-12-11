package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsIamRolePolicyAttachmentImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	role := d.Id()
	conn := meta.(*AWSClient).iamconn
	_, err := conn.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(role),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NoSuchEntity" {
				log.Printf("[WARN] No such entity found for Policy Attachment (%s)", role)
				d.SetId("")
				return []*schema.ResourceData{d}, nil
			}
		}
		return []*schema.ResourceData{d}, err
	}

	args := iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(role),
	}
	results := make([]*schema.ResourceData, 1)
	i := 0
	err = conn.ListAttachedRolePoliciesPages(&args, func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
		for _, p := range page.AttachedPolicies {
			subResource := resourceAwsIamRolePolicyAttachment()
			attachment := subResource.Data(nil)
			attachment.SetType("aws_iam_role_policy_attachment")
			attachment.Set("role", role)
			attachment.Set("policy_arn", aws.StringValue(p.PolicyArn))
			attachment.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", role)))
			results[i] = attachment
			i++
		}
		return !lastPage
	})
	return results, err
}
