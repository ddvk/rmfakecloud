import 'bootstrap/dist/css/bootstrap.min.css'
import React from 'react';
import Nav from 'react-bootstrap/Nav'
import Row from 'react-bootstrap/Row'
import Navbar from 'react-bootstrap/Navbar'
import Container from 'react-bootstrap/Container'

function App() {
    return (
        <>
        <Navbar bg="dark" variant="dark" >
            <Navbar.Brand href="#home">Stuff</Navbar.Brand>
            <Nav activeKey="/home" onSelect={(k)=> console.log(k)} className="mr-auto" >
            <Nav.Item>
            <Nav.Link href="/home">Active</Nav.Link>
            </Nav.Item>
        </Nav>
        </Navbar>
        <Container fluid={true}>
            <Row>
            <Nav activeKey="/home" onSelect={(k)=> console.log(k)} className="flex-column" >
                <Nav.Item>
                    <Nav.Link href="/home">Active</Nav.Link>
                </Nav.Item>
                <Nav.Item>
                    <Nav.Link href="/home">Active</Nav.Link>
                </Nav.Item>
            </Nav>
            <div className="flex-column">
                <h1>some content</h1>
            </div>
        </Row>
        </Container>
        </>
    );
}

export default App;
