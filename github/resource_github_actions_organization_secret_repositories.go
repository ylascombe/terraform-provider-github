package github

import (
	"context"

	"github.com/google/go-github/v66/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubActionsOrganizationSecretRepositories() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubActionsOrganizationSecretRepositoriesCreateOrUpdate,
		Read:   resourceGithubActionsOrganizationSecretRepositoriesRead,
		Update: resourceGithubActionsOrganizationSecretRepositoriesCreateOrUpdate,
		Delete: resourceGithubActionsOrganizationSecretRepositoriesDelete,
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
			"selected_repository_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Set:         schema.HashInt,
				Required:    true,
				Description: "An array of repository ids that can access the organization secret.",
			},
		},
	}
}

func resourceGithubActionsOrganizationSecretRepositoriesCreateOrUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	secretName := d.Get("secret_name").(string)
	selectedRepositories := d.Get("selected_repository_ids")

	selectedRepositoryIDs := []int64{}

	ids := selectedRepositories.(*schema.Set).List()
	for _, id := range ids {
		selectedRepositoryIDs = append(selectedRepositoryIDs, int64(id.(int)))
	}

	_, err = client.Actions.SetSelectedReposForOrgSecret(ctx, owner, secretName, selectedRepositoryIDs)
	if err != nil {
		return err
	}

	d.SetId(secretName)
	return resourceGithubActionsOrganizationSecretRepositoriesRead(d, meta)
}

func resourceGithubActionsOrganizationSecretRepositoriesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.Background()

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	repositoryID := d.Get("repository_id").(int64)
	var found bool

	opt := &github.ListOptions{
		PerPage: maxPerPage,
	}
	for {
		results, resp, err := client.Actions.ListSelectedReposForOrgSecret(ctx, owner, d.Id(), opt)
		if err != nil {
			return err
		}

		for _, repo := range results.Repositories {
			if repo.GetID() == repositoryID {
				if err := d.Set("selected_repository_ids", []int64{repositoryID}); err != nil {
					return err
				}
				found = true
				break
			}
		}
		if found || resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	if !found {
		if err := d.Set("selected_repository_ids", []int64{}); err != nil {
			return err
		}
	}

	return nil
}

func resourceGithubActionsOrganizationSecretRepositoriesDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	ctx := context.WithValue(context.Background(), ctxId, d.Id())

	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	var selectedRepositoryID int64
	_, err = client.Actions.SetSelectedRepoForOrgSecret(ctx, owner, d.Id(), selectedRepositoryID)
	if err != nil {
		return err
	}

	return nil
}
