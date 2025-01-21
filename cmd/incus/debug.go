package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	cli "github.com/lxc/incus/v6/internal/cmd"
	"github.com/lxc/incus/v6/internal/i18n"
	"github.com/spf13/cobra"
)

type cmdDebug struct {
	global *cmdGlobal
}

func (c *cmdDebug) Command() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = usage("debug")
	cmd.Short = i18n.G("Debug commands")
	cmd.Long = cli.FormatSection(i18n.G("Description"), i18n.G(
		`Debug commands for instances`))

	debugAttachCmd := cmdDebugMemory{global: c.global, debug: c}
	cmd.AddCommand(debugAttachCmd.Command())

	return cmd
}

type cmdDebugMemory struct {
	global *cmdGlobal
	debug  *cmdDebug

	flagFormat string
}

func (c *cmdDebugMemory) Command() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = usage("get-instance-memory", i18n.G("[<remote>:]<instance> [target] [--format]"))
	cmd.Short = i18n.G("Export a virtual machine's memory state")
	cmd.Long = cli.FormatSection(i18n.G("Description"), i18n.G(
		`Export the current memory state of a running virtual machine into a dump file. 
		This can be useful for debugging or analysis purposes.`))
	cmd.Example = cli.FormatSection("", i18n.G(
		`incus debug get-instance-memory vm1 dump.elf --format=elf
    Creates an ELF format memory dump of the vm1 instance.`))

	cmd.RunE = c.Run
	cmd.Flags().StringVar(&c.flagFormat, "format", "elf", i18n.G("Format of memory dump (elf, win-dmp, or kdump formats with compression: kdump-zlib/lzo/snappy, kdump-raw-zlib/lzo/snappy)"))

	return cmd
}

func (c *cmdDebugMemory) Run(cmd *cobra.Command, args []string) error {
	conf := c.global.conf

	// Quick checks.
	exit, err := c.global.CheckArgs(cmd, args, 2, 2)
	if exit {
		return err
	}

	// Connect to the daemon
	remote, name, err := conf.ParseRemote(args[0])
	if err != nil {
		return err
	}

	d, err := conf.GetInstanceServer(remote)
	if err != nil {
		return err
	}

	format := c.flagFormat
	path := args[1]
	ext := strings.ToLower(filepath.Ext(path))

	validFormats := []string{
		"elf",
		"kdump-zlib",
		"kdump-lzo",
		"kdump-snappy",
		"kdump-raw-zlib",
		"kdump-raw-lzo",
		"kdump-raw-snappy",
		"win-dmp",
	}

	if !slices.Contains(validFormats, format) {
		return fmt.Errorf(i18n.G("Invalid format '%s'. Supported formats: %s"), format, strings.Join(validFormats, ", "))
	}

	switch {
	case format == "elf" && ext != ".elf":
		return fmt.Errorf(i18n.G("Format 'elf' requires .elf file extension"))
	case format == "win-dmp" && ext != ".dmp":
		return fmt.Errorf(i18n.G("Format 'win-dmp' requires .dmp file extension"))
	case strings.HasPrefix(format, "kdump-") && ext != ".dump":
		return fmt.Errorf(i18n.G("Kdump formats require .dump file extension"))
	}

	err = d.GetInstanceDebugMemory(name, path, format)
	if err != nil {
		return fmt.Errorf(i18n.G("Failed to dump instance memory: %w"), err)
	}

	// Watch the background operation
	progress := cli.ProgressRenderer{
		Format: i18n.G("Exporting VM memory: %s"),
		Quiet:  c.global.flagQuiet,
	}

	progress.Done(i18n.G("Memory dump completed successfully!"))

	return nil
}
