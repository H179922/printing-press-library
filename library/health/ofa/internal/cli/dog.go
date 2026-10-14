// Copyright 2026 H179922 and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live

package cli

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/health/ofa/internal/ofa"
)

var appnumArgRe = regexp.MustCompile(`^\d{1,10}$`)

func newDogCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dog [appnum]",
		Short: "Show one dog's published OFA record: conclusions, sire, dam, and relatives",
		Long: `Show one dog's full published OFA record by application number (the
appnum column in 'results'). Includes every registry conclusion with test and
report dates, the sire and dam registration numbers, and related dogs OFA lists.`,
		Example:     "  ofa-pp-cli dog 2682765\n  ofa-pp-cli dog 2682765 --json --select tests",
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": "<appnum>=2682765"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if len(args) != 1 || !appnumArgRe.MatchString(args[0]) {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("exactly one numeric appnum is required (see the appnum column in 'results')"))
			}
			appnum := args[0]
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "fetch OFA record "+appnum)
			}
			c, err := newOFAClient(flags)
			if err != nil {
				return err
			}
			body, err := c.Get(cmd.Context(), ofa.SearchPath, map[string]string{"a": ofa.DetailRoute, "appnum": appnum})
			if err != nil {
				return classifyAPIErrorOnly(fmt.Errorf("OFA record %s: %w", appnum, err))
			}
			d, err := ofa.ParseDog(body, appnum)
			if errors.Is(err, ofa.ErrDogNotFound) {
				return notFoundErr(fmt.Errorf("no OFA record for appnum %s", appnum))
			}
			if err != nil {
				return apiErr(err)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printDogHuman(cmd, d)
			}
			return printJSONFiltered(cmd.OutOrStdout(), d, flags)
		},
	}
	return cmd
}

func printDogHuman(cmd *cobra.Command, d ofa.Dog) error {
	w := cmd.OutOrStdout()
	chic := ""
	if d.CHIC {
		chic = "  [CHIC]"
	}
	fmt.Fprintf(w, "%s%s\n%s  %s %s %s  born %s\n", d.Name, chic, d.Registration, d.Sex, d.Color, d.Breed, d.Born)
	if d.Sire != "" || d.Dam != "" {
		fmt.Fprintf(w, "Sire: %s   Dam: %s\n", orDash(d.Sire), orDash(d.Dam))
	}
	fmt.Fprintln(w)
	if len(d.Tests) == 0 {
		fmt.Fprintln(w, "No published test results.")
	} else if err := printAutoTable(w, toMaps(d.Tests)); err != nil {
		return err
	}
	if len(d.Relatives) > 0 {
		fmt.Fprintf(w, "\nRelatives (%d):\n", len(d.Relatives))
		return printAutoTable(w, toMaps(d.Relatives))
	}
	return nil
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
