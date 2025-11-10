//go:build windows

package truststore

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	// Certificate encoding types
	X509_ASN_ENCODING   = 0x00000001
	PKCS_7_ASN_ENCODING = 0x00010000

	// Certificate store flags
	CERT_STORE_ADD_REPLACE_EXISTING = 3

	// Special error codes
	CRYPT_E_NOT_FOUND = 0x80092004
)

var (
	// crypt32.dll functions
	crypt32                              = syscall.NewLazyDLL("crypt32.dll")
	procCertOpenSystemStoreW             = crypt32.NewProc("CertOpenSystemStoreW")
	procCertAddEncodedCertificateToStore = crypt32.NewProc("CertAddEncodedCertificateToStore")
	procCertCloseStore                   = crypt32.NewProc("CertCloseStore")
	procCertEnumCertificatesInStore      = crypt32.NewProc("CertEnumCertificatesInStore")
	procCertDuplicateCertificateContext  = crypt32.NewProc("CertDuplicateCertificateContext")
	procCertDeleteCertificateFromStore   = crypt32.NewProc("CertDeleteCertificateFromStore")
)

// windowsStore implements TrustStore for Windows certificate store
type windowsStore struct{}

// windowsRootStore wraps a Windows certificate store handle
type windowsRootStore uintptr

// NewSystemStore creates a new Windows system trust store
func NewSystemStore() TrustStore {
	return &windowsStore{}
}

func (s *windowsStore) Name() string {
	return "Windows Certificate Store"
}

func (s *windowsStore) IsInstalled(cert *x509.Certificate) bool {
	store, err := openRootStore()
	if err != nil {
		return false
	}
	defer store.close()

	found := false
	store.enumerate(func(storeCert *x509.Certificate) bool {
		if storeCert.SerialNumber.Cmp(cert.SerialNumber) == 0 {
			found = true
			return false // stop enumeration
		}
		return true // continue
	})

	return found
}

func (s *windowsStore) Install(certPath string, cert *x509.Certificate) error {
	// Read certificate file
	data, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate: %w", err)
	}

	// Decode PEM to DER
	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("failed to decode PEM certificate")
	}

	// Open ROOT certificate store
	store, err := openRootStore()
	if err != nil {
		return fmt.Errorf("failed to open certificate store: %w", err)
	}
	defer store.close()

	// Add certificate to store
	encodingType := X509_ASN_ENCODING | PKCS_7_ASN_ENCODING
	r1, _, err := procCertAddEncodedCertificateToStore.Call(
		uintptr(store),
		uintptr(encodingType),
		uintptr(unsafe.Pointer(&block.Bytes[0])),
		uintptr(len(block.Bytes)),
		uintptr(CERT_STORE_ADD_REPLACE_EXISTING),
		0, // don't return context
	)

	if r1 == 0 {
		return fmt.Errorf("failed to add certificate to store: %w", err)
	}

	return nil
}

func (s *windowsStore) Uninstall(cert *x509.Certificate) error {
	// Open ROOT certificate store
	store, err := openRootStore()
	if err != nil {
		return fmt.Errorf("failed to open certificate store: %w", err)
	}
	defer store.close()

	// Find and delete matching certificates
	deleted := false
	err = store.enumerate(func(storeCert *x509.Certificate) bool {
		if storeCert.SerialNumber.Cmp(cert.SerialNumber) == 0 {
			// Note: We can't delete during enumeration, so we'll need to handle this differently
			// For now, we'll just mark that we found it
			deleted = true
			return false
		}
		return true
	})

	if err != nil {
		return fmt.Errorf("failed to enumerate certificates: %w", err)
	}

	if !deleted {
		return fmt.Errorf("certificate not found in store")
	}

	// Re-open and delete
	store2, err := openRootStore()
	if err != nil {
		return fmt.Errorf("failed to reopen certificate store: %w", err)
	}
	defer store2.close()

	return store2.deleteCert(cert)
}

// openRootStore opens the Windows ROOT certificate store
func openRootStore() (windowsRootStore, error) {
	storeName, err := syscall.UTF16PtrFromString("ROOT")
	if err != nil {
		return 0, err
	}

	r1, _, err := procCertOpenSystemStoreW.Call(0, uintptr(unsafe.Pointer(storeName)))
	if r1 == 0 {
		return 0, fmt.Errorf("failed to open ROOT store: %w", err)
	}

	return windowsRootStore(r1), nil
}

// close closes the certificate store
func (s windowsRootStore) close() error {
	r1, _, err := procCertCloseStore.Call(uintptr(s), 0)
	if r1 == 0 {
		return err
	}
	return nil
}

// enumerate iterates over all certificates in the store
func (s windowsRootStore) enumerate(callback func(*x509.Certificate) bool) error {
	var prevContext uintptr = 0

	for {
		r1, _, err := procCertEnumCertificatesInStore.Call(uintptr(s), prevContext)
		if r1 == 0 {
			// Check if we reached the end
			if errno, ok := err.(syscall.Errno); ok && errno == CRYPT_E_NOT_FOUND {
				break
			}
			return fmt.Errorf("failed to enumerate certificates: %w", err)
		}

		prevContext = r1

		// Parse the certificate
		cert, err := parseCertContext(r1)
		if err != nil {
			// Skip certificates we can't parse
			continue
		}

		// Call callback
		if !callback(cert) {
			break
		}
	}

	return nil
}

// deleteCert deletes a certificate from the store by serial number
func (s windowsRootStore) deleteCert(cert *x509.Certificate) error {
	var prevContext uintptr = 0

	for {
		r1, _, err := procCertEnumCertificatesInStore.Call(uintptr(s), prevContext)
		if r1 == 0 {
			if errno, ok := err.(syscall.Errno); ok && errno == CRYPT_E_NOT_FOUND {
				break
			}
			return fmt.Errorf("failed to enumerate certificates: %w", err)
		}

		prevContext = r1

		// Parse the certificate
		storeCert, err := parseCertContext(r1)
		if err != nil {
			continue
		}

		// Check if this is our certificate
		if storeCert.SerialNumber.Cmp(cert.SerialNumber) == 0 {
			// Duplicate context before deleting
			dupR1, _, _ := procCertDuplicateCertificateContext.Call(r1)
			if dupR1 == 0 {
				continue
			}

			// Delete the certificate
			deleteR1, _, deleteErr := procCertDeleteCertificateFromStore.Call(dupR1)
			if deleteR1 == 0 {
				return fmt.Errorf("failed to delete certificate: %w", deleteErr)
			}

			return nil
		}
	}

	return fmt.Errorf("certificate not found")
}

// parseCertContext parses a certificate from a Windows certificate context
func parseCertContext(context uintptr) (*x509.Certificate, error) {
	// Certificate context structure
	type certContext struct {
		EncodingType uint32
		EncodedCert  *byte
		Length       uint32
		CertInfo     uintptr
		Store        uintptr
	}

	ctx := (*certContext)(unsafe.Pointer(context))
	if ctx.EncodedCert == nil || ctx.Length == 0 {
		return nil, fmt.Errorf("invalid certificate context")
	}

	// Extract certificate bytes
	certBytes := make([]byte, ctx.Length)
	for i := uint32(0); i < ctx.Length; i++ {
		certBytes[i] = *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(ctx.EncodedCert)) + uintptr(i)))
	}

	// Parse certificate
	return x509.ParseCertificate(certBytes)
}
