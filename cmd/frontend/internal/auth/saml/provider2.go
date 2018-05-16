package saml

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/beevik/etree"
	"github.com/pkg/errors"
	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

var mockGetServiceProvider2 func(*schema.SAMLAuthProvider) (*saml2.SAMLServiceProvider, error)

func getServiceProvider2(ctx context.Context, pc *schema.SAMLAuthProvider) (*saml2.SAMLServiceProvider, error) {
	if mockGetServiceProvider2 != nil {
		return mockGetServiceProvider2(pc)
	}

	c, err := readProviderConfig(pc, conf.Get().AppURL)
	if err != nil {
		return nil, err
	}

	idpMetadata, err := readIdentityProviderMetadata(ctx, c)
	if err != nil {
		return nil, err
	}

	metadata, err := unmarshalEntityDescriptor2(idpMetadata)
	if err != nil {
		return nil, errors.WithMessage(err, "parsing SAML Identity Provider metadata")
	}

	certStore := dsig.MemoryX509CertificateStore{Roots: []*x509.Certificate{}}
	for _, kd := range metadata.IDPSSODescriptor.KeyDescriptors {
		for i, xcert := range kd.KeyInfo.X509Data.X509Certificates {
			if xcert.Data == "" {
				return nil, fmt.Errorf("SAML Identity Provider metadata certificate %d is empty", i)
			}
			certData, err := base64.StdEncoding.DecodeString(xcert.Data)
			if err != nil {
				return nil, errors.WithMessage(err, fmt.Sprintf("decoding SAML Identity Provider metadata certificate %d", i))
			}
			idpCert, err := x509.ParseCertificate(certData)
			if err != nil {
				return nil, errors.WithMessage(err, fmt.Sprintf("parsing SAML Identity Provider metadata certificate %d X.509 data", i))
			}
			certStore.Roots = append(certStore.Roots, idpCert)
		}
	}

	issuerURL := c.entityID.ResolveReference(&url.URL{Path: path.Join(c.entityID.Path, "/saml/metadata")}).String()
	return &saml2.SAMLServiceProvider{
		IdentityProviderSSOURL:      metadata.IDPSSODescriptor.SingleSignOnServices[0].Location,
		IdentityProviderIssuer:      metadata.EntityID,
		ServiceProviderIssuer:       issuerURL,
		AssertionConsumerServiceURL: c.entityID.ResolveReference(&url.URL{Path: path.Join(c.entityID.Path, "/saml/acs")}).String(),
		SignAuthnRequests:           true,
		AudienceURI:                 issuerURL,
		IDPCertificateStore:         &certStore,
		SPKeyStore:                  dsig.TLSCertKeyStore(c.keyPair),
		// TODO(sqs): Use the persistent NameIDFormat (https://github.com/sourcegraph/sourcegraph/issues/11206).
		NameIdFormat:           "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
		ValidateEncryptionCert: true,
	}, nil
}

// entitiesDescriptor2 represents the SAML EntitiesDescriptor object.
//
// It is very similar to github.com/crewjam/saml/samlsp's EntitiesDescriptor, except it uses types
// from github.com/russellhaering/gosaml2 instead (to be compatible with the rest of the new (2)
// impl).
type entitiesDescriptor2 struct {
	XMLName             xml.Name       `xml:"urn:oasis:names:tc:SAML:2.0:metadata EntitiesDescriptor"`
	ID                  *string        `xml:",attr,omitempty"`
	ValidUntil          *time.Time     `xml:"validUntil,attr,omitempty"`
	CacheDuration       *time.Duration `xml:"cacheDuration,attr,omitempty"`
	Name                *string        `xml:",attr,omitempty"`
	Signature           *etree.Element
	EntitiesDescriptors []entitiesDescriptor2    `xml:"urn:oasis:names:tc:SAML:2.0:metadata EntitiesDescriptor"`
	EntityDescriptors   []types.EntityDescriptor `xml:"urn:oasis:names:tc:SAML:2.0:metadata EntityDescriptor"`
}

// unmarshalEntityDescriptor2 unmarshals from an XML root <EntityDescriptor> or <EntitiesDescriptor>
// element. If the latter, it returns the first <EntityDescriptor> child that has an
// IDPSSODescriptor.
//
// Taken from github.com/crewjam/saml. Similar to unmarshalEntityDescriptor1, except it uses types
// for the new (2) impl.
func unmarshalEntityDescriptor2(data []byte) (*types.EntityDescriptor, error) {
	var entity *types.EntityDescriptor
	if err := xml.Unmarshal(data, &entity); err != nil {
		// This comparison is ugly, but it is how the error is generated in encoding/xml.
		if err.Error() != "expected element type <EntityDescriptor> but have <EntitiesDescriptor>" {
			return nil, err
		}
		var entities *entitiesDescriptor2
		if err := xml.Unmarshal(data, &entities); err != nil {
			return nil, err
		}
		for i, e := range entities.EntityDescriptors {
			if e.IDPSSODescriptor != nil {
				entity = &entities.EntityDescriptors[i]
				break
			}
		}
		if entity == nil {
			return nil, errors.New("no entity found with IDPSSODescriptor")
		}
	}
	return entity, nil
}