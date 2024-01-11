// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signoption

import (
	"fmt"
	"strings"

	"github.com/open-component-model/ocm/pkg/signing/signutils"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/keyoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/hashoption"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/normalizations/jsonv1"
	ocmsign "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

var _ options.Options = (*Option)(nil)

func New(sign bool) *Option {
	return &Option{SignMode: sign}
}

type Option struct {
	keyoption.Option

	local bool

	SignMode      bool
	signAlgorithm string
	// Verify the digests
	Verify bool

	// Recursively sign component versions
	Recursively bool
	// SignatureNames is a list of signatures to handle (only the first one
	// will be used for signing
	SignatureNames []string
	Update         bool
	Signer         signing.Signer

	UseTSA bool
	TSAUrl string

	Hash hashoption.Option

	Keyless bool
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.Option.AddFlags(fs)
	fs.StringArrayVarP(&o.SignatureNames, "signature", "s", nil, "signature name")
	if o.SignMode {
		o.Hash.AddFlags(fs)
		fs.StringVarP(&o.signAlgorithm, "algorithm", "S", rsa.Algorithm, "signature handler")
		fs.BoolVarP(&o.Update, "update", "", o.SignMode, "update digest in component versions")
		fs.BoolVarP(&o.Recursively, "recursive", "R", false, "recursively sign component versions")
		fs.BoolVarP(&o.UseTSA, "tsa", "", false, fmt.Sprintf("use timestamp authority (default server: %s)", signing.DEFAULT_TSA_URL))
		fs.StringVarP(&o.TSAUrl, "tsa-url", "", "", "TSA server URL")
	} else {
		fs.BoolVarP(&o.local, "local", "L", false, "verification based on information found in component versions, only")
	}
	fs.BoolVarP(&o.Verify, "verify", "V", o.SignMode, "verify existing digests")
	fs.BoolVar(&o.Keyless, "keyless", false, "use keyless signing")
}

func (o *Option) Configure(ctx clictx.Context) error {
	if len(o.SignatureNames) > 0 {
		for i, n := range o.SignatureNames {
			n = strings.TrimSpace(n)
			dn, err := signutils.ParseDN(n)
			if err != nil {
				return err
			}
			o.SignatureNames[i] = signutils.NormalizeDN(*dn)
			if n == "" {
				return errors.Newf("empty signature name (name %d) not possible", i)
			}
		}
		o.DefaultName = o.SignatureNames[0] // set default name for key handling
	} else {
		o.SignatureNames = nil
	}

	if o.SignMode {
		err := o.Hash.Configure(ctx)
		if err != nil {
			return err
		}
		if o.signAlgorithm == "" {
			o.signAlgorithm = rsa.Algorithm
		}
		o.Signer = signingattr.Get(ctx).GetSigner(o.signAlgorithm)
		if o.Signer == nil {
			return errors.ErrUnknown(compdesc.KIND_SIGN_ALGORITHM, o.signAlgorithm)
		}
	} else {
		o.Recursively = !o.local
	}

	err := o.Option.Configure(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (o *Option) Usage() string {
	s := `
The <code>--public-key</code> and <code>--private-key</code> options can be
used to define public and private keys on the command line. The options have an
argument of the form <code>[&lt;name>=]&lt;filepath></code>. The optional name
specifies the signature name the key should be used for. By default, this is the
signature name specified with the option <code>--signature</code>.

Alternatively a key can be specified as base64 encoded string if the argument
start with the prefix <code>!</code> or as direct string with the prefix
<code>=</code>.
`

	if o.SignMode {
		s += `
If in signing mode a public key is specified, existing signatures for the
given signature name will be verified, instead of recreated.
`
		s += `

The following signing types are supported with option <code>--algorithm</code>:
` + listformat.FormatList(rsa.Algorithm, signing.DefaultRegistry().SignerNames()...)

		s += `

The following normalization modes are supported with option <code>--normalization</code>:
` + listformat.FormatList(jsonv1.Algorithm, compdesc.Normalizations.Names()...)

		s += `

The following hash modes are supported with option <code>--hash</code>:
` + listformat.FormatList(sha256.Algorithm, signing.DefaultRegistry().HasherNames()...)

		signing.DefaultRegistry().HasherNames()
	}
	return s
}

var _ ocmsign.Option = (*Option)(nil)

func (o *Option) ApplySigningOption(opts *ocmsign.Options) {
	if o.Signer != nil {
		opts.Signer = o.Signer
	}
	opts.SignatureNames = o.SignatureNames
	opts.Verify = o.Verify
	opts.Recursively = o.Recursively
	opts.Keys = o.Keys
	opts.NormalizationAlgo = o.Hash.NormAlgorithm
	opts.Hasher = o.Hash.Hasher

	if o.UseTSA || o.TSAUrl != "" {
		opts.UseTSA = true
		if o.TSAUrl != "" {
			opts.TSAUrl = o.TSAUrl
		}
	}
	if o.Keys != nil {
		def := o.Keys.GetIssuer("")
		if def != nil {
			opts.Issuer = def
		}
	}
	if o.RootCerts != nil {
		opts.RootCerts = o.RootCerts
	}
	if len(o.SignatureNames) > 0 {
		if o.Keyless {
			opts.VerifySignature = true
		} else {
			opts.VerifySignature = o.Keys.GetPublicKey(o.SignatureNames[0]) != nil
		}
	}
	opts.Update = o.Update
	opts.Keyless = o.Keyless
}
