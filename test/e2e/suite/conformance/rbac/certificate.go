/*
Copyright 2020 The cert-manager Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rbac

import (
	"context"

	"github.com/cert-manager/cert-manager/e2e-tests/framework"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = RBACDescribe("Certificates", func() {
	f := framework.NewDefaultFramework("rbac-certificates")
	resource := "certificates" // this file is related to certificates

	Context("with namespace view access", func() {
		clusterRole := "view"
		It("shouldn't be able to create certificates", func(testingCtx context.Context) {
			verb := "create"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeFalse())
		})

		It("shouldn't be able to delete certificates", func(testingCtx context.Context) {
			verb := "delete"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeFalse())
		})

		It("shouldn't be able to delete collections of certificates", func(testingCtx context.Context) {
			verb := "deletecollection"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeFalse())
		})

		It("shouldn't be able to patch certificates", func(testingCtx context.Context) {
			verb := "patch"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeFalse())
		})

		It("shouldn't be able to update certificates", func(testingCtx context.Context) {
			verb := "update"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeFalse())
		})

		It("should be able to get certificates", func(testingCtx context.Context) {
			verb := "get"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to list certificates", func(testingCtx context.Context) {
			verb := "list"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to watch certificates", func(testingCtx context.Context) {
			verb := "watch"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})
	})
	Context("with namespace edit access", func() {
		clusterRole := "edit"
		It("should be able to create certificates", func(testingCtx context.Context) {
			verb := "create"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to delete certificates", func(testingCtx context.Context) {
			verb := "delete"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to delete collections of certificates", func(testingCtx context.Context) {
			verb := "deletecollection"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to patch certificates", func(testingCtx context.Context) {
			verb := "patch"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to update certificates", func(testingCtx context.Context) {
			verb := "update"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to get certificates", func(testingCtx context.Context) {
			verb := "get"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to list certificates", func(testingCtx context.Context) {
			verb := "list"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to watch certificates", func(testingCtx context.Context) {
			verb := "watch"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})
	})

	Context("with namespace admin access", func() {
		clusterRole := "admin"
		It("should be able to create certificates", func(testingCtx context.Context) {
			verb := "create"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to delete certificates", func(testingCtx context.Context) {
			verb := "delete"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to delete collections of certificates", func(testingCtx context.Context) {
			verb := "deletecollection"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to patch certificates", func(testingCtx context.Context) {
			verb := "patch"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to update certificates", func(testingCtx context.Context) {
			verb := "update"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to get certificates", func(testingCtx context.Context) {
			verb := "get"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to list certificates", func(testingCtx context.Context) {
			verb := "list"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})

		It("should be able to watch certificates", func(testingCtx context.Context) {
			verb := "watch"

			hasAccess := framework.RbacClusterRoleHasAccessToResource(testingCtx, f, clusterRole, verb, resource)
			Expect(hasAccess).Should(BeTrue())
		})
	})
})
