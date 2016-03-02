package drivers_test

import (
	"light-stemcell-builder/config"
	"light-stemcell-builder/driversets"
	"light-stemcell-builder/resources"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VolumeDriver", func() {
	It("creates an EBS Volume from a previously uploaded machine image", func() {
		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		Expect(accessKey).ToNot(BeEmpty(), "AWS_ACCESS_KEY_ID must be set")

		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		Expect(secretKey).ToNot(BeEmpty(), "AWS_SECRET_ACCESS_KEY must be set")

		region := os.Getenv("AWS_REGION")
		Expect(region).ToNot(BeEmpty(), "AWS_REGION must be set")

		creds := config.Credentials{
			AccessKey: accessKey,
			SecretKey: secretKey,
			Region:    region,
		}

		machineImagePath := os.Getenv("MACHINE_IMAGE_PATH")
		Expect(machineImagePath).ToNot(BeEmpty(), "MACHINE_IMAGE_PATH must be set")

		bucketName := os.Getenv("AWS_BUCKET_NAME")
		Expect(bucketName).ToNot(BeEmpty(), "AWS_BUCKET_NAME must be set")

		ds := driversets.NewIsolatedRegionDriverSet(GinkgoWriter, creds)
		machineImageDriverConfig := resources.MachineImageDriverConfig{
			MachineImagePath: machineImagePath,
			BucketName:       bucketName,
		}

		machineImageDriver := ds.CreateMachineImageDriver()
		machineImage, err := machineImageDriver.Create(machineImageDriverConfig)
		Expect(err).ToNot(HaveOccurred())

		volumeDriverConfig := resources.VolumeDriverConfig{
			MachineImageManifestURL: machineImage.GetURL,
		}

		volumeDriver := ds.CreateVolumeDriver()

		volume, err := volumeDriver.Create(volumeDriverConfig)
		Expect(err).ToNot(HaveOccurred())

		ec2Client := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})
		reqOutput, err := ec2Client.DescribeVolumes(&ec2.DescribeVolumesInput{VolumeIds: []*string{aws.String(volume.ID)}})
		Expect(err).ToNot(HaveOccurred())

		Expect(reqOutput.Volumes).To(HaveLen(1))

		_, err = ec2Client.DeleteVolume(&ec2.DeleteVolumeInput{VolumeId: aws.String(volume.ID)})
		Expect(err).ToNot(HaveOccurred())
	})
})