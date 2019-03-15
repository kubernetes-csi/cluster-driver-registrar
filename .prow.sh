#! /bin/bash -e

# cluster-driver-registrar is not part of any of the current hostpath
# driver deployments, therefore we disable installation and testing of
# those. Instead, the code below runs a custom E2E test suite.
CSI_PROW_DEPLOYMENT=none

. release-tools/prow.sh

# main handles non-E2E testing and cluster installation for us.
if ! main; then
    ret=1
else
    ret=0
fi

if [ "$KUBECONFIG" ]; then
    # We have a cluster. Run our own E2E testing.
    collect_cluster_info
    install_ginkgo
    args=
    if ${CSI_PROW_BUILD_JOB}; then
        # Image was side-loaded into the cluster.
        args=-cluster-driver-registrar-image=csi-cluster-driver-registrar:csiprow
    fi
    if ! run ginkgo -v "./test/e2e" -- -repo-root="$(pwd)" -report-dir "${ARTIFACTS}" $args; then
        warn "e2e suite failed"
        ret=1
    fi
fi

# Merge all junit files into one. This ensures that Spyglass finds them (seems to ignore junit_make_test.xml).
if ls "${ARTIFACTS}"/junit_*.xml 2>/dev/null >&2; then
    run_filter_junit -o "${CSI_PROW_WORK}/junit_final.xml" "${ARTIFACTS}"/junit_*.xml && rm "${ARTIFACTS}"/junit_*.xml && mv "${CSI_PROW_WORK}/junit_final.xml" "${ARTIFACTS}"
fi

exit $ret
