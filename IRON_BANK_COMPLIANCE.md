# Iron Bank Compliance Report

## Document Scanner Service (DWS) - Iron Bank Acceptance Baseline Criteria Compliance

This document outlines the compliance status of the DWS application with Iron Bank Acceptance Baseline Criteria v1.1.

## ✅ COMPLIANT REQUIREMENTS

### 4.2.1 Fundamental Requirements
- [x] **Multi-stage build**: Uses UBI8 builder → Distroless runtime
- [x] **No internet downloads**: All dependencies vendored in `vendor/` directory
- [x] **Entrypoint scripts**: Located in `scripts/entrypoint.sh` as required
- [x] **No development tools**: Final distroless image contains only runtime essentials
- [x] **Installation binaries removed**: Multi-stage build removes build artifacts

### 4.2.2 Scanning Requirements
- [x] **No encrypted content**: Container content is not encrypted
- [x] **Virus scanning**: Ready for Iron Bank pipeline virus scanning
- [x] **Scanner policies**: No circumvention of scanning tools

### 4.2.3 Documentation Requirements
- [x] **README.md**: Comprehensive documentation in markdown format
- [x] **LICENSE**: Apache 2.0 license included
- [x] **Documentation folder**: Documentation copied to `/documentation/` folder
- [x] **OCI Labels**: Defined in `hardening_manifest.yaml`

### 4.2.4 Iron Bank Compliance Requirements
- [x] **Base Images**:
  - Builder: `registry1.dso.mil/ironbank/redhat/ubi/ubi8` ✅
  - Runtime: `registry1.dso.mil/ironbank/google/distroless/static` ✅
- [x] **STDOUT logging**: All logs sent to STDOUT
- [x] **TLS Support**: Application supports TLS 1.2/1.3 when configured
- [x] **Non-root execution**: Distroless runs as nobody user (UID 65534)
- [x] **File permissions**: No SUID/GUID bits, secure permissions

### 4.2.5 Support Requirements
- [x] **Active maintenance**: Application is actively developed
- [x] **Open source**: Apache 2.0 licensed open source software

### Future Requirements (4.2.6)
- [ ] **FIPS compliance**: Not yet implemented (planned for future release)

## 📋 IRON BANK ARTIFACTS CREATED

### Required Files
1. **`hardening_manifest.yaml`** - Complete hardening manifest with:
   - Repository and maintainer information
   - Resource checksums and dependencies
   - OCI labels and metadata
   - Security context configuration
   - Build arguments and environment variables

2. **`scripts/entrypoint.sh`** - Iron Bank compliant entrypoint script:
   - Located in required `/scripts/` folder
   - Executable permissions set
   - Handles health checks and configuration
   - Logs to STDOUT as required

3. **`documentation/README.md`** - Documentation in required folder

4. **`LICENSE`** - Apache 2.0 license (already existed)

### Docker Image Compliance
- **FROM images**: Only Iron Bank approved base images
- **Build process**: No internet downloads during build
- **Final image**: Distroless for minimal attack surface
- **User context**: Non-root execution
- **Health checks**: Implemented as required

## 🔒 SECURITY IMPROVEMENTS

### CVE Mitigation
- **Go 1.24**: Latest stable version for security patches
- **Distroless base**: Minimal attack surface with no shell/package manager
- **Non-root execution**: Runs as nobody user (UID 65534)
- **Vendor dependencies**: All dependencies vendored and pinned

### File Permissions
- No SUID/GUID bits set
- No world-writable files
- Standard secure Unix permissions

## 📊 VULNERABILITY COMPLIANCE

Per Iron Bank Table B requirements:
- **Critical (9.0-10.0)**: Must remediate within 15 days, max 1 finding
- **High (7.0-8.9)**: Must remediate within 35 days, max 4 findings
- **Medium (4.0-6.9)**: Must remediate within 180 days
- **Low (0.1-3.9)**: Must remediate within 360 days

**Current Status**: Ready for Iron Bank vulnerability scanning pipeline

## 🚀 DEPLOYMENT READY

The DWS application is now compliant with Iron Bank Acceptance Baseline Criteria and ready for:

1. **Iron Bank Submission**: All required artifacts present
2. **Vulnerability Scanning**: Compliant container structure
3. **Security Review**: Hardening standards implemented
4. **Production Deployment**: Kubernetes manifests included

## 📝 NEXT STEPS

1. Submit to Iron Bank repository pipeline
2. Complete vulnerability scan and remediation
3. Obtain Iron Bank approval status
4. Deploy using provided Kubernetes manifests

---

**Compliance Date**: $(date -u +%Y-%m-%d)
**Iron Bank Criteria Version**: 1.1
**DWS Version**: 1.0.0