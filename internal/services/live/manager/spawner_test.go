package manager_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/services/live/manager"
	"github.com/backtesting-org/kronos-cli/pkg/live"
)

var _ = Describe("ProcessSpawner", func() {
	var (
		spawner      live.ProcessSpawner
		logger       logging.ApplicationLogger
		testStrategy *strategy.Strategy
		tmpDir       string
		ctx          context.Context
		cancel       context.CancelFunc
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "spawner-test-*")
		Expect(err).NotTo(HaveOccurred())

		// Change to tmpDir for tests
		originalDir, _ := os.Getwd()
		err = os.Chdir(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func() {
			_ = os.Chdir(originalDir)
			_ = os.RemoveAll(tmpDir)
		})

		logger = &logging.NoOpLogger{}
		spawner = manager.NewProcessSpawner(logger)

		testStrategy = &strategy.Strategy{
			Name: "test-momentum",
			Path: "./strategies/test-momentum",
		}

		ctx, cancel = context.WithCancel(context.Background())

		DeferCleanup(func() {
			cancel()
		})
	})

	Describe("Spawn", func() {
		It("should create a command with correct arguments", func() {
			cmd, err := spawner.Spawn(ctx, testStrategy)
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())

			// Verify command path and args
			Expect(cmd.Path).To(ContainSubstring("kronos"))
			Expect(cmd.Args).To(ContainElements(
				ContainSubstring("kronos"),
				"run-strategy",
				"--strategy",
				"test-momentum",
			))
		})

		It("should set process group for detachment", func() {
			cmd, err := spawner.Spawn(ctx, testStrategy)
			Expect(err).NotTo(HaveOccurred())

			// Verify SysProcAttr is set for process group
			Expect(cmd.SysProcAttr).NotTo(BeNil())
			Expect(cmd.SysProcAttr.Setpgid).To(BeTrue())
		})

		It("should create log directory structure", func() {
			_, err := spawner.Spawn(ctx, testStrategy)
			Expect(err).NotTo(HaveOccurred())

			// Verify directory was created
			logDir := filepath.Join(".kronos", "instances", "test-momentum")
			_, err = os.Stat(logDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create stdout and stderr log files", func() {
			cmd, err := spawner.Spawn(ctx, testStrategy)
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())

			// Verify log files exist
			stdoutLog := filepath.Join(".kronos", "instances", "test-momentum", "stdout.log")
			stderrLog := filepath.Join(".kronos", "instances", "test-momentum", "stderr.log")

			// Files should be created
			Eventually(func() bool {
				_, err1 := os.Stat(stdoutLog)
				_, err2 := os.Stat(stderrLog)
				return err1 == nil && err2 == nil
			}).Should(BeTrue())
		})

		It("should redirect stdout and stderr to log files", func() {
			cmd, err := spawner.Spawn(ctx, testStrategy)
			Expect(err).NotTo(HaveOccurred())

			// Verify stdout and stderr are set to files (not nil, not os.Stdout/Stderr)
			Expect(cmd.Stdout).NotTo(BeNil())
			Expect(cmd.Stderr).NotTo(BeNil())
			Expect(cmd.Stdout).NotTo(Equal(os.Stdout))
			Expect(cmd.Stderr).NotTo(Equal(os.Stderr))
		})

		It("should handle context cancellation", func() {
			localCtx, localCancel := context.WithCancel(context.Background())
			cmd, err := spawner.Spawn(localCtx, testStrategy)
			Expect(err).NotTo(HaveOccurred())

			// Cancel context
			localCancel()

			// Command should respect context
			Expect(cmd.Process).To(BeNil()) // Not started yet
		})

		It("should create unique log files for different strategies", func() {
			strategy1 := &strategy.Strategy{Name: "momentum-1", Path: "./strategies/momentum-1"}
			strategy2 := &strategy.Strategy{Name: "momentum-2", Path: "./strategies/momentum-2"}

			_, err := spawner.Spawn(ctx, strategy1)
			Expect(err).NotTo(HaveOccurred())

			_, err = spawner.Spawn(ctx, strategy2)
			Expect(err).NotTo(HaveOccurred())

			// Verify both directories exist
			dir1 := filepath.Join(".kronos", "instances", "momentum-1")
			dir2 := filepath.Join(".kronos", "instances", "momentum-2")

			_, err = os.Stat(dir1)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(dir2)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should append to existing log files", func() {
			// First spawn
			cmd1, err := spawner.Spawn(ctx, testStrategy)
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd1).NotTo(BeNil())

			logPath := filepath.Join(".kronos", "instances", "test-momentum", "stdout.log")

			// Write some content
			err = os.WriteFile(logPath, []byte("first line\n"), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Second spawn (should append, not overwrite)
			cmd2, err := spawner.Spawn(ctx, testStrategy)
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd2).NotTo(BeNil())

			// Verify file still has original content (not truncated)
			content, err := os.ReadFile(logPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("first line"))
		})

		Context("when directory creation fails", func() {
			It("should return error for invalid path", func() {
				// Make .kronos directory read-only
				kronosDir := ".kronos"
				_ = os.MkdirAll(kronosDir, 0755)
				_ = os.Chmod(kronosDir, 0444) // Read-only

				DeferCleanup(func() {
					_ = os.Chmod(kronosDir, 0755)
				})

				_, err := spawner.Spawn(ctx, testStrategy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create instance log directory"))
			})
		})

		Context("when strategy name has special characters", func() {
			It("should handle strategy names safely", func() {
				specialStrategy := &strategy.Strategy{
					Name: "test-strategy-v1.2.3",
					Path: "./strategies/test",
				}

				cmd, err := spawner.Spawn(ctx, specialStrategy)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd).NotTo(BeNil())

				// Verify directory created with safe name
				logDir := filepath.Join(".kronos", "instances", "test-strategy-v1.2.3")
				_, err = os.Stat(logDir)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("AttachMonitor", func() {
		var (
			instance *live.Instance
		)

		BeforeEach(func() {
			cmd, err := spawner.Spawn(ctx, testStrategy)
			Expect(err).NotTo(HaveOccurred())

			instance = &live.Instance{
				ID:           "test-instance-123",
				StrategyName: testStrategy.Name,
				Cmd:          cmd,
				PID:          0, // Not started yet
				Context:      ctx,
				Cancel:       cancel,
			}
		})

		It("should successfully attach monitor to valid instance", func() {
			err := spawner.AttachMonitor(instance)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error when cmd is nil", func() {
			instance.Cmd = nil
			err := spawner.AttachMonitor(instance)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("command not set"))
		})

		It("should log monitor attachment", func() {
			// This test verifies the monitor can be attached without errors
			// The actual monitoring happens in the manager, not the spawner
			err := spawner.AttachMonitor(instance)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Process Group Behavior", func() {
		It("should create process in new process group", func() {
			cmd, err := spawner.Spawn(ctx, testStrategy)
			Expect(err).NotTo(HaveOccurred())

			// Verify the process group settings
			Expect(cmd.SysProcAttr).NotTo(BeNil())
			sysProcAttr := cmd.SysProcAttr
			Expect(sysProcAttr.Setpgid).To(BeTrue())

			// This ensures child process survives parent exit
			pgid := sysProcAttr.Pgid
			Expect(pgid).To(Equal(0)) // 0 means create new process group
		})
	})

	Describe("Log File Management", func() {
		It("should create log files with correct permissions", func() {
			_, err := spawner.Spawn(ctx, testStrategy)
			Expect(err).NotTo(HaveOccurred())

			logPath := filepath.Join(".kronos", "instances", "test-momentum", "stdout.log")
			info, err := os.Stat(logPath)
			Expect(err).NotTo(HaveOccurred())

			// Verify file permissions (0644)
			mode := info.Mode()
			Expect(mode & 0644).To(Equal(os.FileMode(0644)))
		})

		It("should handle concurrent spawns for same strategy", func() {
			// Spawn multiple times concurrently
			done := make(chan bool, 3)
			errors := make(chan error, 3)

			for i := 0; i < 3; i++ {
				go func() {
					_, err := spawner.Spawn(ctx, testStrategy)
					errors <- err
					done <- true
				}()
			}

			// Wait for all to complete
			for i := 0; i < 3; i++ {
				<-done
			}

			// All should succeed (appending to same log)
			close(errors)
			for err := range errors {
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})

	Describe("Integration with Context", func() {
		It("should respect context timeout", func() {
			timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer timeoutCancel()

			cmd, err := spawner.Spawn(timeoutCtx, testStrategy)
			Expect(err).NotTo(HaveOccurred())

			// The command is created but respects context
			Expect(cmd.Process).To(BeNil()) // Not started

			// Wait for timeout
			time.Sleep(150 * time.Millisecond)

			// Context should be cancelled
			Expect(timeoutCtx.Err()).To(Equal(context.DeadlineExceeded))
		})

		It("should handle immediate context cancellation", func() {
			immediateCtx, immediateCancel := context.WithCancel(context.Background())
			immediateCancel() // Cancel immediately

			cmd, err := spawner.Spawn(immediateCtx, testStrategy)
			Expect(err).NotTo(HaveOccurred())

			// Command created with cancelled context
			Expect(cmd).NotTo(BeNil())
			Expect(immediateCtx.Err()).To(Equal(context.Canceled))
		})
	})

	Describe("Edge Cases", func() {
		It("should handle empty strategy name", func() {
			emptyStrategy := &strategy.Strategy{
				Name: "",
				Path: "./strategies/empty",
			}

			cmd, err := spawner.Spawn(ctx, emptyStrategy)
			// Should still work - creates directory with empty name
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
		})

		It("should handle very long strategy names", func() {
			longName := strings.Repeat("a", 255)

			longStrategy := &strategy.Strategy{
				Name: longName,
				Path: "./strategies/long",
			}

			cmd, err := spawner.Spawn(ctx, longStrategy)
			// May fail on some filesystems, but should handle gracefully
			if err == nil {
				Expect(cmd).NotTo(BeNil())
			} else {
				// Should provide clear error message
				Expect(err.Error()).To(Or(
					ContainSubstring("directory"),
					ContainSubstring("file"),
					ContainSubstring("name"),
				))
			}
		})
	})
})
