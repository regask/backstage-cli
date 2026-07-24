package cmd

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/regask/backstage-regask-cli/internal/az"
	"github.com/regask/backstage-regask-cli/internal/contracts"
	"github.com/regask/backstage-regask-cli/internal/render"
	"github.com/spf13/cobra"
)

var (
	secretsEnv    string
	secretsReveal bool
	secretsVault  string
)

type secretRow struct {
	SecretKey string `json:"secretKey"`
	VaultKey  string `json:"vaultKey"`
	Value     string `json:"value,omitempty"`
}

var checkSecretsCmd = &cobra.Command{
	Use:   "check-secrets <service>",
	Short: "Show a service's secret refs (values resolved via `az`, masked by default)",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		if secretsEnv == "" {
			return fmt.Errorf("--env is required")
		}
		cl, err := newClient()
		if err != nil {
			return err
		}
		ctx := context.Background()
		ov, err := cl.Overlays(ctx, args[0], freshFlag)
		if err != nil {
			return err
		}
		refs := contracts.MergeMaps(
			contracts.ParseSecretRefs(ov.SecretBase),
			contracts.ParseSecretRefs(ov.SecretOverlay[secretsEnv]),
		)
		vault := secretsVault
		if vault == "" {
			vault = ov.Vaults[secretsEnv]
		}

		runner := az.NewCLI()
		rows := make([]secretRow, 0, len(refs))
		for _, key := range contracts.SortedKeys(refs) {
			row := secretRow{SecretKey: key, VaultKey: refs[key]}
			if secretsReveal {
				if vault == "" {
					return fmt.Errorf("no vault for env %q; pass --vault", secretsEnv)
				}
				val, err := runner.GetSecret(ctx, vault, refs[key])
				if err != nil {
					return err
				}
				row.Value = val
			} else {
				row.Value = "********"
			}
			rows = append(rows, row)
		}
		return render.Output(JSONOutput(), rows, func(w io.Writer) {
			tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
			fmt.Fprintln(tw, "SECRET\tVAULT KEY\tVALUE")
			for _, r := range rows {
				fmt.Fprintf(tw, "%s\t%s\t%s\n", r.SecretKey, r.VaultKey, r.Value)
			}
			tw.Flush()
		})
	},
}

func init() {
	checkSecretsCmd.Flags().StringVar(&secretsEnv, "env", "", "environment (required)")
	checkSecretsCmd.Flags().BoolVar(&secretsReveal, "reveal", false, "resolve and print secret values via az")
	checkSecretsCmd.Flags().StringVar(&secretsVault, "vault", "", "override Azure Key Vault name")
	checkSecretsCmd.Flags().BoolVar(&freshFlag, "fresh", false, "bypass server cache")
	RootCmd.AddCommand(checkSecretsCmd)
}
