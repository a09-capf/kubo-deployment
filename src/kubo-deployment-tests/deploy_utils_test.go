package kubo_deployment_tests_test

import (
	"fmt"

	. "github.com/jhvhs/gob-mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"path"
)

var _ = Describe("DeployUtils", func() {
	Describe("set_cloud_config", func() {
		It("Should generate cloud config", func() {
			bash.Source(pathToScript("lib/deploy_utils"), nil)

			bash.Export("BOSH_ENV", "kubo-env")
			bash.Source("", func(string) ([]byte, error) {
				return []byte(fmt.Sprintf("export PATH=%s:$PATH", pathFromRoot("bin"))), nil
			})

			generateCloudConfigMock := Mock("generate_cloud_config", `echo -n "cc"`)
			boshCliMock := Mock("bosh-cli", `echo -n "$@"`)
			ApplyMocks(bash, []Gob{generateCloudConfigMock, boshCliMock})

			code, err := bash.Run("set_cloud_config", []string{})
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0))
			Expect(stderr).To(gbytes.Say("generate_cloud_config kubo-env"))
		})

		It("Should update cloud config", func() {
			bash.Source(pathToScript("lib/deploy_utils"), nil)

			bash.Export("BOSH_ENV", "kubo-env")
			bash.Export("BOSH_NAME", "env-name")
			bash.Source("", func(string) ([]byte, error) {
				return []byte(fmt.Sprintf("export PATH=%s:$PATH", pathFromRoot("bin"))), nil
			})

			generateCloudConfigMock := Mock("generate_cloud_config", `echo -n "cc"`)
			boshCliMock := Mock("bosh-cli", `echo -n "$@"`)
			ApplyMocks(bash, []Gob{generateCloudConfigMock, boshCliMock})

			code, err := bash.Run("set_cloud_config", []string{})
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0))
			Expect(stderr).To(gbytes.Say("bosh-cli -n -e env-name update-cloud-config -"))
		})
	})

	Describe("export_bosh_environment", func() {
		It("should set BOSH_ENV and BOSH_NAME", func() {
			bash.Source(pathToScript("lib/deploy_utils"), nil)
			code, err := bash.Run("export_bosh_environment /envs/foo && echo $BOSH_ENV && echo $BOSH_NAME", []string{})
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0))
			Expect(stdout).To(gbytes.Say("/envs/foo"))
			Expect(stdout).To(gbytes.Say("foo"))
		})
	})

	Describe("deploy_to_bosh", func() {
		It("should deploy to bosh", func() {
			bash.Source(pathToScript("lib/deploy_utils"), nil)

			getBoshSecretMock := Mock("get_bosh_secret", `echo "the-secret"`)
			boshCliMock := Mock("bosh-cli", `echo -n "$@"`)
			ApplyMocks(bash, []Gob{ getBoshSecretMock, boshCliMock })

			bash.Export("BOSH_ENV", "kubo-env")
			bash.Export("BOSH_NAME", "env-name")

			code, err := bash.Run("deploy_to_bosh", []string{"manifest", "deployment-name"})
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0))
			Expect(stderr).To(gbytes.Say("bosh-cli -d deployment-name -n deploy --no-redact -"))
		})
	})

	Describe("get_bosh_secret", func() {
		It("should get bosh_admin_client_secret setting", func() {
			bash.Source(pathToScript("lib/deploy_utils"), nil)

			getSettingMock := Mock("get_setting", `echo "the-secret"`)
			ApplyMocks(bash, []Gob{ getSettingMock })

			code, err := bash.Run("get_bosh_secret", []string{})
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0))
			Expect(stdout).To(gbytes.Say("the-secret"))
		})
	})

	Describe("get_setting", func() {
		It("should call bosh interpolate", func() {
			bash.Source(pathToScript("lib/deploy_utils"), nil)

			boshCliMock := Mock("bosh-cli", `echo -n "$@"`)
			ApplyMocks(bash, []Gob{ boshCliMock })

			bash.Export("BOSH_ENV", "kubo-env")

			code, err := bash.Run("get_setting", []string{"fileToQuery.yml", "path/subpath"})
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0))
			Expect(stderr).To(gbytes.Say("bosh-cli int kubo-env/fileToQuery.yml --path path/subpath"))
		})

		It("should return the setting value", func() {
			bash.Source(pathToScript("lib/deploy_utils"), nil)

			boshCliMock := Mock("bosh-cli", `echo "value-at-path"`)
			ApplyMocks(bash, []Gob{ boshCliMock })

			code, err := bash.Run("get_setting", []string{"fileToQuery.yml", "path/subpath"})
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0))
			Expect(stdout).To(gbytes.Say("value-at-path"))
		})
	})

	Describe("create_and_upload_release", func(){
		Context("is called with a valid directory path", func(){
			It("should create a release and upload it", func(){
				bash.Source(pathToScript("lib/deploy_utils"), nil)

				getBoshSecretMock := Mock("get_bosh_secret", `echo "the-secret"`)
				boshCliMock := Mock("bosh-cli", `echo -n "$@"`)
				uploadReleaseMock := Mock("upload_release", `echo`)
				ApplyMocks(bash, []Gob{ getBoshSecretMock, boshCliMock, uploadReleaseMock})


				releasePath := path.Join(resourcesPath, "releases/mock-release")

				code, err := bash.Run("create_and_upload_release", []string{releasePath})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0))
				Expect(stderr).To(gbytes.Say("bosh-cli create-release --force --name mock"))
				Expect(stderr).To(gbytes.Say("upload_release --name=mock"))
			})
		})

		Context("is called with an invalid argument", func(){
			It("should exit", func(){
				bash.Source(pathToScript("lib/deploy_utils"), nil)
				code, err := bash.Run("create_and_upload_release", []string{"path_does_not_exist"})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(1))
			})
		})
	})

	Describe("upload_release", func(){
		It("should upload release", func() {
			bash.Source(pathToScript("lib/deploy_utils"), nil)

			getBoshSecretMock := Mock("get_bosh_secret", `echo "the-secret"`)
			boshCliMock := Mock("bosh-cli", `echo -n "$@"`)
			ApplyMocks(bash, []Gob{getBoshSecretMock, boshCliMock})

			bash.Export("BOSH_ENV", "kubo-env")
			bash.Export("BOSH_NAME", "env-name")

			code, err := bash.Run("upload_release", []string{"release-name"})
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0))
			Expect(stderr).To(gbytes.Say("bosh-cli upload-release release-name"))
		})
	})

	Describe( "set_ops_file_if_path_exists", func() {
		Context("when the variable path exists", func() {
			It("returns an ops-file argument", func() {
				bash.Source(pathToScript("lib/deploy_utils"), nil)

				boshMock := Mock("bosh-cli", `return 0`)
				ApplyMocks(bash, []Gob{boshMock})

				code, err := bash.Run("set_ops_file_if_path_exists",
					[]string{
						"director.yml",
						"/some_variable",
						"some-ops-file.yml",
						})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0))
				Expect(stdout).To(gbytes.Say("some-ops-file.yml"))
			})
		})

		Context("when the variable path doesn't exist", func() {
			It("returns an empty string", func() {
				bash.Source(pathToScript("lib/deploy_utils"), nil)

				boshMock := Mock("bosh-cli", `return 1`)
				ApplyMocks(bash, []Gob{boshMock})

				code, err := bash.Run("set_ops_file_if_path_exists",
					[]string{
						"director.yml",
						"/some_variable",
						"some-ops-file.yml",
					})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0))
				Expect(stdout).NotTo(gbytes.Say("some-ops-file.yml"))
			})
		})
	})

	Describe("set_ops_file_if_file_exists", func() {
		Context("when the file exists", func() {
			It("returns an ops-file path", func() {
				bash.Source(pathToScript("lib/deploy_utils"), nil)

				code, err := bash.Run("set_ops_file_if_file_exists",
					[]string{
						pathFromRoot("manifests/ops-files/iaas/aws/cloud-provider.yml"),
					})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0))
				Expect(stdout).To(gbytes.Say("cloud-provider.yml"))
			})
		})

		Context("when the file doesn't exist", func() {
			It("returns an empty string", func() {
				bash.Source(pathToScript("lib/deploy_utils"), nil)

				code, err := bash.Run("set_ops_file_if_file_exists",
					[]string{
						"a-file-which-does-not-exist.yml",
					})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0))
				Expect(stdout).NotTo(gbytes.Say("a-file-which-does-not-exist.yml"))
			})
		})
	})

	Describe("set_vars_file_if_file_exists", func() {
		Context("when the file exists", func() {
			It("returns a vars-file path", func() {
				bash.Source(pathToScript("lib/deploy_utils"), nil)

				code, err := bash.Run("set_vars_file_if_file_exists",
					[]string{
						path.Join(testEnvironmentPath, "with_vars/name-vars.yml"),
					})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0))
				Expect(stdout).To(gbytes.Say("name-vars.yml"))
			})
		})

		Context("when the file doesn't exist", func() {
			It("returns an empty string", func() {
				bash.Source(pathToScript("lib/deploy_utils"), nil)

				code, err := bash.Run("set_vars_file_if_file_exists",
					[]string{
						"a-file-which-does-not-exist.yml",
					})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0))
				Expect(stdout).NotTo(gbytes.Say("a-file-which-does-not-exist.yml"))
			})
		})
	})

	Describe("set_default_var_if_path_does_not_exist", func() {
		Context("when the variable path doesn't exist", func() {
			It("returns a variable with default value", func() {
				bash.Source(pathToScript("lib/deploy_utils"), nil)

				boshMock := Mock("bosh-cli", `return 1`)
				ApplyMocks(bash, []Gob{boshMock})

				code, err := bash.Run("set_default_var_if_path_does_not_exist",
					[]string{
						"director.yml",
						"/some_variable",
						"default-value",
					})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0))
				Expect(stdout).To(gbytes.Say("--var some_variable=default-value"))
			})
		})

		Context("when the variable path exists", func() {
			It("returns an empty string", func() {
				bash.Source(pathToScript("lib/deploy_utils"), nil)

				boshMock := Mock("bosh-cli", `echo "some-value"`)
				ApplyMocks(bash, []Gob{boshMock})

				code, err := bash.Run("set_default_var_if_path_does_not_exist",
					[]string{
						"director.yml",
						"/some_variable",
						"default-value",
					})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0))
				Expect(stdout).NotTo(gbytes.Say("--var some_variable=default-value"))
			})
		})

	})

	Describe("generate_manifest", func() {
		BeforeEach(func() {
			bash.Source(pathToScript("lib/deploy_utils"), nil)
		})

		It("applies dev, bootstrap and use-runtime-config-bosh-dns ops files", func() {
			boshMock := Mock("bosh-cli", `
			if [[ "$3" =~ "addons_spec_path" ]]; then
				return 1
			elif [[ "$3" =~ "routing_mode" ]]; then
				echo "the-routing-mode"
			elif [[ "$3" =~ "iaas" ]]; then
				echo "vsphere"
			else
				echo
			fi`)
			setOpsFileIfPathExistsMock := Mock("set_ops_file_if_path_exists",
				`echo --ops-file="$3"`)
			setOpsFileIfFileExistsMock := Mock("set_ops_file_if_file_exists",
				`echo --ops-file="$1"`)
			setVarsFileIfFileExistsMock := Mock("set_vars_file_if_file_exists",
				`echo --vars-file="$1"`)
			setDefaultVarIfPathDoesNotExistMock := Mock("set_default_var_if_path_does_not_exist",
				`echo --var ${2#/}=$3`)
			ApplyMocks(bash, []Gob{
				boshMock,
				setOpsFileIfPathExistsMock,
				setOpsFileIfFileExistsMock,
				setVarsFileIfFileExistsMock,
				setDefaultVarIfPathDoesNotExistMock,
			})

			code, err := bash.Run("generate_manifest", []string{
				path.Join(testEnvironmentPath, "with_ops_and_vars_and_creds"),
				"name",
				pathFromRoot("manifests/cfcr.yml"),
				"director-uuid"})
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(0))

			Expect(stderr).To(gbytes.Say("routing_mode"))
			Expect(stderr).To(gbytes.Say("iaas"))

			Expect(stderr).To(gbytes.Say(fmt.Sprintf("%s %s %s %s",
				"set_ops_file_if_path_exists",
				path.Join(testEnvironmentPath, "with_ops_and_vars_and_creds/director.yml"),
				"/http_proxy",
				pathFromRoot("manifests/ops-files/add-http-proxy.yml"),
			)))
			Expect(stderr).To(gbytes.Say(fmt.Sprintf("%s %s %s %s",
				"set_ops_file_if_path_exists",
				path.Join(testEnvironmentPath, "with_ops_and_vars_and_creds/director.yml"),
				"/https_proxy",
				pathFromRoot("manifests/ops-files/add-https-proxy.yml"),
			)))
			Expect(stderr).To(gbytes.Say(fmt.Sprintf("%s %s %s %s",
				"set_ops_file_if_path_exists",
				path.Join(testEnvironmentPath, "with_ops_and_vars_and_creds/director.yml"),
				"/no_proxy",
				pathFromRoot("manifests/ops-files/add-no-proxy.yml"),
			)))
			Expect(stderr).To(gbytes.Say(fmt.Sprintf("%s %s",
				"set_ops_file_if_file_exists",
				pathFromRoot("manifests/ops-files/iaas/vsphere/cloud-provider.yml"),
			)))
			Expect(stderr).To(gbytes.Say(fmt.Sprintf("%s %s",
				"set_ops_file_if_file_exists",
				path.Join(testEnvironmentPath, "with_ops_and_vars_and_creds/name.yml"),
			)))
			Expect(stderr).To(gbytes.Say(fmt.Sprintf("%s %s",
				"set_vars_file_if_file_exists",
				path.Join(testEnvironmentPath, "with_ops_and_vars_and_creds/name-vars.yml"),
			)))
			Expect(stderr).To(gbytes.Say(fmt.Sprintf("%s %s",
				"set_vars_file_if_file_exists",
				path.Join(testEnvironmentPath, "with_ops_and_vars_and_creds/creds.yml"),
			)))
			Expect(stderr).To(gbytes.Say(fmt.Sprintf("%s %s",
				"set_vars_file_if_file_exists",
				path.Join(testEnvironmentPath, "with_ops_and_vars_and_creds/director-secrets.yml"),
			)))
			Expect(stderr).To(gbytes.Say(fmt.Sprintf("%s %s %s %s",
				"set_default_var_if_path_does_not_exist",
				path.Join(testEnvironmentPath, "with_ops_and_vars_and_creds/director.yml"),
				"/authorization_mode",
				"abac",
			)))
			Expect(stderr).To(gbytes.Say(fmt.Sprintf("%s %s %s %s",
				"set_default_var_if_path_does_not_exist",
				path.Join(testEnvironmentPath, "with_ops_and_vars_and_creds/director.yml"),
				"/worker_count",
				"3",
			)))

			Expect(stderr).To(gbytes.Say("misc/dev.yml"))
			Expect(stderr).To(gbytes.Say("misc/bootstrap.yml"))
			Expect(stderr).To(gbytes.Say("use-runtime-config-bosh-dns.yml"))
			Expect(stderr).To(gbytes.Say("director.yml"))
			Expect(stderr).To(gbytes.Say("--var deployment_name=name"))
			Expect(stderr).To(gbytes.Say("--var director_uuid=director-uuid"))
			Expect(stderr).To(gbytes.Say("add-http-proxy.yml"))
			Expect(stderr).To(gbytes.Say("add-https-proxy.yml"))
			Expect(stderr).To(gbytes.Say("add-no-proxy.yml"))
			Expect(stderr).To(gbytes.Say("cloud-provider.yml"))
			Expect(stderr).To(gbytes.Say("name.yml"))
			Expect(stderr).To(gbytes.Say("name-vars.yml"))
			Expect(stderr).To(gbytes.Say("creds.yml"))
			Expect(stderr).To(gbytes.Say("director-secrets.yml"))
			Expect(stderr).To(gbytes.Say("--var authorization_mode=abac"))
			Expect(stderr).To(gbytes.Say("--var worker_count=3"))
		})

		Context("when routing_mode is cf", func() {
			It("applies cf-routing ops-file", func() {
				boshMock := Mock("bosh-cli", `
				if [[ "$3" =~ "addons_spec_path" \
					|| "$3" =~ "http_proxy" \
					|| "$3" =~ "https_proxy" \
					|| "$3" =~ "no_proxy" ]]; then
					return 1
				elif [[ "$3" =~ "routing_mode" ]]; then
					echo "cf"
				elif [[ "$3" =~ "iaas" ]]; then
					echo "the-iaas"
				else
					echo
				fi`)
				ApplyMocks(bash, []Gob{boshMock})

				code, err := bash.Run("generate_manifest", []string{"environment-path", "deployment-name", "non-existent-manifest-path", "director-uuid"})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0))
				Expect(stderr).To(gbytes.Say("cf-routing.yml"))
			})
		})

		Context("when iaas is aws", func() {
			It("applies aws lb ops-file", func() {
				boshMock := Mock("bosh-cli", `
				if [[ "$3" =~ "addons_spec_path" \
					|| "$3" =~ "http_proxy" \
					|| "$3" =~ "https_proxy" \
					|| "$3" =~ "no_proxy" ]]; then
					return 1
				elif [[ "$3" =~ "routing_mode" ]]; then
					echo "the-routing-mode"
				elif [[ "$3" =~ "iaas" ]]; then
					echo "aws"
				else
					echo
				fi`)
				ApplyMocks(bash, []Gob{boshMock})

				code, err := bash.Run("generate_manifest", []string{"environment-path", "deployment-name", "non-existent-manifest-path", "director-uuid"})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0))
				Expect(stderr).To(gbytes.Say("aws/lb.yml"))
			})
		})

		Context("when iaas is gcp", func() {
			Context("when service_account_worker is not set", func(){
				It("applies the add-service-key-worker ops-file", func() {
					boshMock := Mock("bosh-cli", `
						if [[ "$3" =~ "addons_spec_path" \
							|| "$3" =~ "http_proxy" \
							|| "$3" =~ "https_proxy" \
							|| "$3" =~ "no_proxy" \
							|| "$3" =~ "service_account_worker" ]]; then
							return 1
						elif [[ "$3" =~ "routing_mode" ]]; then
							echo "the-routing-mode"
						elif [[ "$3" =~ "iaas" ]]; then
							echo "gcp"
						else
							echo
						fi`)
					ApplyMocks(bash, []Gob{boshMock})

					code, err := bash.Run("generate_manifest", []string{"environment-path", "deployment-name", "non-existent-manifest-path", "director-uuid"})
					Expect(err).NotTo(HaveOccurred())
					Expect(code).To(Equal(0))
					Expect(stderr).To(gbytes.Say("gcp/add-service-key-worker.yml"))
				})
			})

			Context("when service_account_master is not set", func(){
				It("applies the add-service-key-master ops-file", func() {
					boshMock := Mock("bosh-cli", `
						if [[ "$3" =~ "addons_spec_path" \
							|| "$3" =~ "http_proxy" \
							|| "$3" =~ "https_proxy" \
							|| "$3" =~ "no_proxy" \
							|| "$3" =~ "service_account_master" ]]; then
							return 1
						elif [[ "$3" =~ "routing_mode" ]]; then
							echo "the-routing-mode"
						elif [[ "$3" =~ "iaas" ]]; then
							echo "gcp"
						else
							echo
						fi`)
					ApplyMocks(bash, []Gob{boshMock})

					code, err := bash.Run("generate_manifest", []string{"environment-path", "deployment-name", "non-existent-manifest-path", "director-uuid"})
					Expect(err).NotTo(HaveOccurred())
					Expect(code).To(Equal(0))
					Expect(stderr).To(gbytes.Say("gcp/add-service-key-master.yml"))
				})
			})
		})

		Context("when addons_spec_path exists", func() {
			It("applies addons-spec.yml ops-file and addon_path as vars file", func(){
				boshMock := Mock("bosh-cli", `
					if [[ "$3" =~ "http_proxy" \
						|| "$3" =~ "https_proxy" \
						|| "$3" =~ "no_proxy" ]]; then
						return 1
					elif [[ "$3" =~ "routing_mode" ]]; then
						echo "the-routing-mode"
					elif [[ "$3" =~ "iaas" ]]; then
						echo "the-iaas"
					elif [[ "$3" =~ "addons_spec_path" ]]; then
						echo "addon.yml"
					else
						echo
					fi`)
				ApplyMocks(bash, []Gob{boshMock})

				code, err := bash.Run("generate_manifest", []string{path.Join(testEnvironmentPath, "with_addons"), "deployment-name", pathFromRoot("manifests/cfcr.yml"), "director-uuid"})
				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(0))
				Expect(stderr).To(gbytes.Say("addons-spec.yml"))
				Expect(stderr).To(gbytes.Say("var-file=\"addons-spec"))
				Expect(stderr).To(gbytes.Say("addon.yml"))
			})
		})
	})
})
