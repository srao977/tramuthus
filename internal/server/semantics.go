package server

import (
	"context"

	finfeedsatv1 "fin_feedsat_1/gen/fin_feedsat/v1"
	"fin_feedsat_1/internal/semantics"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) SetSemantics(dict *semantics.Dictionary) {
	s.semantics = dict
}

func (s *Server) GetTerm(_ context.Context, req *finfeedsatv1.GetSemanticTermRequest) (*finfeedsatv1.SemanticTerm, error) {
	dict, err := s.requireSemantics()
	if err != nil {
		return nil, err
	}
	id := req.GetId()
	term, ok := dict.Term(id)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "unknown semantic id %s", id)
	}
	return toProtoSemanticTerm(term, dict.Version()), nil
}

func (s *Server) ListTerms(_ context.Context, req *finfeedsatv1.ListSemanticTermsRequest) (*finfeedsatv1.ListSemanticTermsResponse, error) {
	dict, err := s.requireSemantics()
	if err != nil {
		return nil, err
	}
	terms := dict.List(req.GetComponent(), req.GetType())
	out := make([]*finfeedsatv1.SemanticTerm, 0, len(terms))
	for _, term := range terms {
		out = append(out, toProtoSemanticTerm(term, dict.Version()))
	}
	return &finfeedsatv1.ListSemanticTermsResponse{
		Contract: toProtoSemanticContract(dict),
		Terms:    out,
	}, nil
}

func (s *Server) GetSemanticContract(context.Context, *finfeedsatv1.GetSemanticContractRequest) (*finfeedsatv1.SemanticContractInfo, error) {
	dict, err := s.requireSemantics()
	if err != nil {
		return nil, err
	}
	return toProtoSemanticContract(dict), nil
}

func (s *Server) requireSemantics() (*semantics.Dictionary, error) {
	if s.semantics == nil {
		return nil, status.Error(codes.Unavailable, "semantic dictionary unavailable")
	}
	return s.semantics, nil
}

func toProtoSemanticContract(dict *semantics.Dictionary) *finfeedsatv1.SemanticContractInfo {
	c := dict.Contract()
	return &finfeedsatv1.SemanticContractInfo{
		Name:              c.Name,
		Version:           c.Version,
		Date:              c.Date,
		Status:            c.Status,
		TermCount:         uint32(dict.TermCount()),
		PersistencePolicy: c.PersistencePolicy,
	}
}

func toProtoSemanticTerm(term semantics.Term, version string) *finfeedsatv1.SemanticTerm {
	return &finfeedsatv1.SemanticTerm{
		Id:                      term.ID,
		Term:                    term.Term,
		Display:                 term.Display,
		Type:                    term.Type,
		Component:               term.Component,
		PlainMeaning:            term.PlainMeaning,
		ScientificMeaning:       term.ScientificMeaning,
		Interpretation:          term.Interpretation,
		DoesNotMean:             append([]string(nil), term.DoesNotMean...),
		RelatedTerms:            append([]string(nil), term.RelatedTerms...),
		Tooltip:                 term.UI.Tooltip,
		PopoverTitle:            term.UI.PopoverTitle,
		PopoverBody:             term.UI.PopoverBody,
		ShowScientificDetail:    term.UI.ShowScientificDetail,
		SemanticContractVersion: version,
		GoSymbol:                term.Source.GoSymbol,
		ProtoEnumOrField:        term.Source.ProtoEnumOrField,
		PresentationOnly:        term.PresentationOnly,
		LifecycleStatus:         term.LifecycleStatus,
		CanonicalSourceIds:      append([]string(nil), term.CanonicalSourceIDs...),
		CompatibilityAliases:    append([]string(nil), term.CompatibilityAliases...),
		LivePathProven:          term.LivePathProven,
		PersistencePolicy:       term.PersistencePolicy,
	}
}
