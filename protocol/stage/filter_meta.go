package stage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/imulab/go-scim/core"
	"math/rand"
	"strings"
	"time"
)

const (
	pathMetaResourceType = "meta.resourceType"
	pathMetaCreated      = "meta.created"
	pathMetaLastModified = "meta.lastModified"
	pathMetaLocation     = "meta.location"
	pathMetaVersion      = "meta.version"
)

var (
	_ PropertyFilter = (*metaResourceTypeFilter)(nil)
	_ PropertyFilter = (*metaCreatedFilter)(nil)
	_ PropertyFilter = (*metaLastModifiedFilter)(nil)
	_ PropertyFilter = (*metaLocationFilter)(nil)
	_ PropertyFilter = (*metaVersionFilter)(nil)
)

type (
	metaResourceTypeFilter struct{}
	metaCreatedFilter      struct{}
	metaLastModifiedFilter struct{}
	metaVersionFilter      struct{}
	metaLocationFilter     struct {
		// Map of resource type's id to url template format
		locationFormats map[string]string
	}
)

// Create a meta resource type filter. The filter is responsible of assigning resource's resource type to the field
// 'meta.resourceType'. The filter only assigns the resource type when Filter is called. The filter is a no-op when
// FilterWithRef is called.
func NewMetaResourceTypeFilter() PropertyFilter {
	return &metaResourceTypeFilter{}
}

func (f *metaResourceTypeFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Id == pathMetaResourceType
}

func (f *metaResourceTypeFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaResourceTypeFilter) FilterOnCreate(ctx context.Context,
	resource *core.Resource, property core.Property) error {
	return property.(core.Crud).Replace(nil, resource.GetResourceType().Name)
}

func (f *metaResourceTypeFilter) FilterOnUpdate(ctx context.Context,
	resource *core.Resource, property core.Property,
	ref *core.Resource, refProp core.Property) error {
	return nil
}

// Create a meta created filter. The filter is responsible of assigning the current time to the field 'meta.created'
// when Filter is called. The filter is a no-op when FilterWithRef is called.
func NewMetaCreatedFilter() PropertyFilter {
	return &metaCreatedFilter{}
}

func (f *metaCreatedFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Id == pathMetaCreated
}

func (f *metaCreatedFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaCreatedFilter) FilterOnCreate(ctx context.Context,
	resource *core.Resource, property core.Property) error {
	return property.(core.Crud).Replace(nil, time.Now().Format(core.ISO8601))
}

func (f *metaCreatedFilter) FilterOnUpdate(ctx context.Context,
	resource *core.Resource, property core.Property,
	ref *core.Resource, refProp core.Property) error {
	return nil
}

// Create a meta lastModified filter. The filter is responsible of assigning the current time to the field 'meta.lastModified'
// when either Filter or FilterWithRef is called.
func NewMetaLastModifiedFilter() PropertyFilter {
	return &metaLastModifiedFilter{}
}

func (f *metaLastModifiedFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Id == pathMetaLastModified
}

func (f *metaLastModifiedFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaLastModifiedFilter) FilterOnCreate(ctx context.Context,
	resource *core.Resource, property core.Property) error {
	return property.(core.Crud).Replace(nil, time.Now().Format(core.ISO8601))
}

func (f *metaLastModifiedFilter) FilterOnUpdate(ctx context.Context,
	resource *core.Resource, property core.Property,
	ref *core.Resource, refProp core.Property) error {
	return property.(core.Crud).Replace(nil, time.Now().Format(core.ISO8601))
}

// Create a meta location filter. The filter is responsible of generating the resource location url and assign it to field
// 'meta.location'. Id must have been generated and bulkId is not accepted. Generation only happens when Filter is called;
// when FilterWithRef is called, this is a no-op.
func NewMetaLocationFilter(locationFormats map[string]string) PropertyFilter {
	return &metaLocationFilter{
		locationFormats: locationFormats,
	}
}

func (f *metaLocationFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Id == pathMetaLocation
}

func (f *metaLocationFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaLocationFilter) FilterOnCreate(ctx context.Context,
	resource *core.Resource, property core.Property) error {
	id, err := resource.GetID()
	if err != nil {
		return core.Errors.Internal("failed to obtain resource id")
	} else if strings.HasPrefix(id, "bulkId:") {
		return core.Errors.Internal("location filter failed: cannot process bulkId")
	}

	format := f.locationFormats[resource.GetResourceType().Id]
	if len(format) == 0 {
		panic("location url formats for all resource types must be set in metaFilter")
	}

	return property.(core.Crud).Replace(nil, fmt.Sprintf(format, id))
}

func (f *metaLocationFilter) FilterOnUpdate(ctx context.Context,
	resource *core.Resource, property core.Property,
	ref *core.Resource, refProp core.Property) error {
	return nil
}

// Create a meta version filter. The filter is responsible of assigning a new version based on an sha256 hash of the
// resource's id, current time and a random number in range of [0, 10000). Naturally, id must have been generated.
// The version assignment happens on both Filter and FilterWithRef call.
func NewMetaVersionFilter() PropertyFilter {
	return &metaVersionFilter{}
}

func (f *metaVersionFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Id == pathMetaVersion
}

func (f *metaVersionFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaVersionFilter) FilterOnCreate(ctx context.Context,
	resource *core.Resource, property core.Property) error {
	return f.assignNewVersion(resource, property)
}

func (f *metaVersionFilter) FilterOnUpdate(ctx context.Context,
	resource *core.Resource, property core.Property,
	ref *core.Resource, refProp core.Property) error {
	return f.assignNewVersion(resource, property)
}

func (f *metaVersionFilter) assignNewVersion(resource *core.Resource, property core.Property) error {
	id, err := resource.GetID()
	if err != nil {
		return core.Errors.Internal("failed to obtain resource id")
	}

	sha := sha256.New()
	sum := sha.Sum([]byte(fmt.Sprintf("%s:%d:%d", id, time.Now().Unix(), rand.Intn(10000))))

	return property.(core.Crud).Replace(nil, fmt.Sprintf("W/\"%x\"", sum))
}
