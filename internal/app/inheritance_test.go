package app_test

import (
	"testing"

	"github.com/evanrichards/nestjs-module-lint/internal/app"
)

func TestInheritanceDetection(t *testing.T) {
	tests := []struct {
		name               string
		sourceCode         string
		expectedClass      string
		expectedBaseClass  string
		expectedHasConstructor bool
	}{
		{
			name: "simple inheritance without constructor",
			sourceCode: `
import { Injectable } from '@nestjs/common';
import { BaseService } from './base.service';

@Injectable()
export class UserService extends BaseService {
  someMethod() {
    return 'test';
  }
}`,
			expectedClass:      "UserService",
			expectedBaseClass:  "BaseService",
			expectedHasConstructor: false,
		},
		{
			name: "inheritance with constructor",
			sourceCode: `
import { Injectable } from '@nestjs/common';
import { BaseService } from './base.service';
import { CacheService } from './cache.service';

@Injectable()
export class UserService extends BaseService {
  constructor(private cache: CacheService) {
    super();
  }
}`,
			expectedClass:      "UserService",
			expectedBaseClass:  "BaseService",
			expectedHasConstructor: true,
		},
		{
			name: "no inheritance",
			sourceCode: `
import { Injectable } from '@nestjs/common';
import { DatabaseService } from './database.service';

@Injectable()
export class UserService {
  constructor(private db: DatabaseService) {}
}`,
			expectedClass:      "UserService",
			expectedBaseClass:  "",
			expectedHasConstructor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inheritance, err := app.AnalyzeClassInheritance([]byte(tt.sourceCode))
			if err != nil {
				t.Fatalf("Failed to analyze inheritance: %v", err)
			}

			if len(inheritance) == 0 && tt.expectedBaseClass != "" {
				t.Fatalf("Expected inheritance info but got none")
			}

			if len(inheritance) > 0 && tt.expectedBaseClass == "" {
				t.Fatalf("Expected no inheritance but got: %v", inheritance)
			}

			if len(inheritance) > 0 {
				info := inheritance[0]
				if info.ClassName != tt.expectedClass {
					t.Errorf("Expected class name %s, got %s", tt.expectedClass, info.ClassName)
				}
				if info.BaseClass != tt.expectedBaseClass {
					t.Errorf("Expected base class %s, got %s", tt.expectedBaseClass, info.BaseClass)
				}
				if info.HasConstructor != tt.expectedHasConstructor {
					t.Errorf("Expected hasConstructor %v, got %v", tt.expectedHasConstructor, info.HasConstructor)
				}
			}
		})
	}
}