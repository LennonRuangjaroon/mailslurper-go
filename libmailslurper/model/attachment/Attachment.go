// Copyright 2013-2014 Adam Presley. All rights reserved
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package attachment

type Attachment struct {
	Headers  *AttachmentHeader
	Contents string
}

func NewAttachment(headers *AttachmentHeader, contents string) *Attachment {
	return &Attachment{
		Headers:  headers,
		Contents: contents,
	}
}