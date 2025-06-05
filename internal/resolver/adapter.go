package resolver

import "github.com/evanrichards/nestjs-module-lint/internal/analysis"

// PathResolverAdapter adapts TsPathResolver to implement the analysis.PathResolver interface
type PathResolverAdapter struct {
	resolver *TsPathResolver
}

// NewPathResolverAdapter creates a new path resolver adapter
func NewPathResolverAdapter(resolver *TsPathResolver) analysis.PathResolver {
	return &PathResolverAdapter{
		resolver: resolver,
	}
}

// ResolveImportPath implements the analysis.PathResolver interface
func (p *PathResolverAdapter) ResolveImportPath(baseDir, importPath string) string {
	return p.resolver.ResolveImportPath(baseDir, importPath)
}
