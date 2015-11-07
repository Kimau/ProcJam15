package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"testing"
)

const testData = string(`<?xml version="1.0" encoding="UTF-8"?>
<rdf:RDF xml:base="http://sw.opencyc.org/concept/"
    xmlns="http://sw.opencyc.org/concept/"
    xmlns:owl="http://www.w3.org/2002/07/owl#"
    xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
    xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#"
    xmlns:skos="http://www.w3.org/2004/02/skos/core#"
    xmlns:xsd="http://www.w3.org/2001/XMLSchema#"
    xmlns:cyc="http://sw.cyc.com/"
    xmlns:opencyc="http://sw.opencyc.org/"
    xmlns:cycAnnot="http://sw.cyc.com/CycAnnotations_v1#">

    <owl:Class rdf:about="Mx8Ngh4rlbasWsxLTdy46teZvW-XVh4rvVjuY5wpEbGdrcN5Y29ycA">
        <rdfs:label xml:lang="en">abnormal cell</rdfs:label>
        <Mx4rwLSVCpwpEbGdrcN5Y29ycA xml:lang="en">abnormal biological cell</Mx4rwLSVCpwpEbGdrcN5Y29ycA>
        <Mx4rwLSVCpwpEbGdrcN5Y29ycA xml:lang="en">abnormal biological cells</Mx4rwLSVCpwpEbGdrcN5Y29ycA>
        <Mx4rwLSVCpwpEbGdrcN5Y29ycA xml:lang="en">abnormal cells</Mx4rwLSVCpwpEbGdrcN5Y29ycA>
        <cycAnnot:label xml:lang="en">(AbnormalFn Cell)</cycAnnot:label>
        <rdf:type rdf:resource="Mx4rHIBS0h_TEdaAAABQ2rksLw"/>
        <rdfs:subClassOf rdf:resource="Mx4rv0har5wpEbGdrcN5Y29ycA"/>
        <rdfs:subClassOf rdf:resource="Mx4rvVjuY5wpEbGdrcN5Y29ycA"/>
        <rdfs:subClassOf rdf:resource="Mx4rvVjaApwpEbGdrcN5Y29ycA"/>
        <Mx4rBVVEokNxEdaAAACgydogAg rdf:resource="Mx4ricybQEpWEduAAAACs2IKfQ"/>
        <owl:sameAs rdf:resource="http://sw.cyc.com/concept/Mx8Ngh4rlbasWsxLTdy46teZvW-XVh4rvVjuY5wpEbGdrcN5Y29ycA"/>
    </owl:Class>

    <owl:Class rdf:about="Mx8Ngh4rlbasWsxLTdy46teZvW-XVh4rvVjuY5wpEbGdrcN5Y29ycA">
        <rdfs:label xml:lang="en">abnormal cell</rdfs:label>
        <Mx4rwLSVCpwpEbGdrcN5Y29ycA xml:lang="en">abnormal biological cell</Mx4rwLSVCpwpEbGdrcN5Y29ycA>
        <Mx4rwLSVCpwpEbGdrcN5Y29ycA xml:lang="en">abnormal biological cells</Mx4rwLSVCpwpEbGdrcN5Y29ycA>
        <Mx4rwLSVCpwpEbGdrcN5Y29ycA xml:lang="en">abnormal cells</Mx4rwLSVCpwpEbGdrcN5Y29ycA>
        <cycAnnot:label xml:lang="en">(AbnormalFn Cell)</cycAnnot:label>
        <rdf:type rdf:resource="Mx4rHIBS0h_TEdaAAABQ2rksLw"/>
        <rdfs:subClassOf rdf:resource="Mx4rv0har5wpEbGdrcN5Y29ycA"/>
        <rdfs:subClassOf rdf:resource="Mx4rvVjuY5wpEbGdrcN5Y29ycA"/>
        <rdfs:subClassOf rdf:resource="Mx4rvVjaApwpEbGdrcN5Y29ycA"/>
        <Mx4rBVVEokNxEdaAAACgydogAg rdf:resource="Mx4ricybQEpWEduAAAACs2IKfQ"/>
        <owl:sameAs rdf:resource="http://sw.cyc.com/concept/Mx8Ngh4rlbasWsxLTdy46teZvW-XVh4rvVjuY5wpEbGdrcN5Y29ycA"/>
    </owl:Class>

</rdf:RDF>
`)

func TestXMLParse(t *testing.T) {
	{
		var v OwlData
		xml.Unmarshal([]byte(testData), &v)

		for _, x := range v.ClassList {
			log.Println(">> ", x.Name)
		}
	}

	{
		var v OwlData

		fBytes, _ := ioutil.ReadFile("owl-export-unversioned.owl")
		xml.Unmarshal(fBytes, v)

		for _, x := range v.ClassList {
			log.Println(">> ", x.Name)
		}
	}

}
