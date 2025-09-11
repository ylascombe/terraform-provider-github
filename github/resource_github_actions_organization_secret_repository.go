package github

import (
	"context"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubActionsOrganizationSecretRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubActionsOrganizationSecretRepositoryCreateOrUpdate,
		Read:   resourceGithubActionsOrganizationSecretRepositoryRead,
		Update: resourceGithubActionsOrganizationSecretRepositoryCreateOrUpdate,
		Delete: resourceGithubActionsOrganizationSecretRepositoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"secret_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "Name of the existing secret.",
				ValidateDiagFunc: validateSecretNameFunc,
			},
			"selected_repository_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The repository ID that can access the organization secret.",
			},
		},
	}
}

func resourceGithubActionsOrganizationSecretRepositoryCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	secretName := d.Get("secret_name").(string)
	selectedRepository := d.Get("selected_repository_id")

	selectedRepositoryID := int64(selectedRepository.(int))

	_, err = client.Actions.SetSelectedRepoForOrgSecret(ctx, owner, secretName, selectedRepositoryID)

	if err != nil {
		return err
	}

	d.SetId(secretName)
	return resourceGithubActionsOrganizationSecretRepositoryRead(d, meta)
}

func resourceGithubActionsOrganizationSecretRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	var selectedRepositoryID int64
	opt := &github.ListOptions{
		PerPage: maxPerPage,
	}
	for {
		results, resp, err := client.Actions.ListSelectedReposForOrgSecret(ctx, owner, d.Id(), opt)
		if err != nil {
			return err
		}

		for _, repo := range results.Repositories {
			selectedRepositoryID = append(selectedRepositoryIDs, repo.GetID())
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	if err = d.Set("selected_repository_ids", selectedRepositoryIDs); err != nil {
		return err
	}

	return nil
}

func resourceGithubActionsOrganizationSecretRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	selectedRepositoryIDs := []int64{}
	_, err = client.Actions.SetSelectedReposForOrgSecret(ctx, owner, d.Id(), selectedRepositoryIDs)
	if err != nil {
		return err
	}

	return nil
}
