import React from 'react';

const Footer = () => {
    return (
        <footer className="footer col-12">
            <div className="footer-container">
                <div className="row list-container">
                    <div className="col-10 offset-0 col-lg-4 offset-lg-1 company">
                        <h2 className="logo">MORENT</h2>
                        <p className="vision">
                            Our vision is to provide convenience<br /> and help increase your sales business.
                        </p>
                    </div>
                    <div className="col-6 offset-0 col-lg-2 offset-lg-1">
                        <h3>About</h3>
                        <ul>
                            <li><a href="#">How it works</a></li>
                            <li><a href="#">Featured</a></li>
                            <li><a href="#">Partnership</a></li>
                            <li><a href="#">Business Relation</a></li>
                        </ul>
                    </div>
                    <div className="col-6 col-lg-2">
                        <h3>Community</h3>
                        <ul>
                            <li><a href="#">Events</a></li>
                            <li><a href="#">Blog</a></li>
                            <li><a href="#">Podcast</a></li>
                            <li><a href="#">Invite a friend</a></li>
                        </ul>
                    </div>
                    <div className="col-6 col-lg-2">
                        <h3>Socials</h3>
                        <ul>
                            <li><a href="#">Discord</a></li>
                            <li><a href="#">Instagram</a></li>
                            <li><a href="#">Twitter</a></li>
                            <li><a href="#">Facebook</a></li>
                        </ul>
                    </div>
                </div>
                <hr className="footer-divider" />
                <div className="row copyright-bottom">
                    <div className="col-4 offset-1 o">
                        <p className="copyright">©2022 MORENT. All rights reserved</p>
                    </div>
                    <div className="col-4 offset-2">
                        <ul className="policy">
                            <li><a href="#">Privacy &amp; Policy</a></li>
                            <li><a href="#">Terms &amp; Condition</a></li>
                        </ul>
                    </div>
                </div>
            </div>
        </footer>
    );
};

export default Footer;
