import React, { useState } from 'react';
import { Dropdown, DropdownToggle, DropdownMenu, DropdownItem } from 'reactstrap';

const DashboardDropDownLinks = (props) => {
    const [dropdownOpen, setDropdownOpen] = useState(false);
    const {linkBuilder, linkList, currentUser} = props;

    const toggle = () => setDropdownOpen(prevState => !prevState);

    const claims = linkList.split(",")

    return (
        <Dropdown isOpen={dropdownOpen} toggle={toggle}>
            <DropdownToggle caret>
                Functions for {currentUser}
            </DropdownToggle>
            <DropdownMenu>
                {
                    claims.map((value) => {
                        if (value === currentUser) {return }
                        return <DropdownItem href={linkBuilder(value)}>Functions for {value}</DropdownItem>
                    })
                }
            </DropdownMenu>
        </Dropdown>
    );
}

export { DashboardDropDownLinks };