// Frames (Executor and Validator) for Skupper 1.x
//
// This will be renamed to f2skupper1 in the near future.
//
// Code organization:
//
// - Each operation (TokenCreate, ServiceStatus, LinkDelete, etc..) has its own
// file (token_create, service_status, link_delete).
// - On it, the main 'SkupperOp' and all interface-specific frames are
// implemented
// - Each interface-specific frame also has the full code for all Skupper versions
// it supports
package f2skupper1
