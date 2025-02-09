import React from "react";

function Table({ data }) {
    if (data.length === 0) {
    return <p>Нет доступных данных.</p>;
}

return (
<table className="ping-table">
<thead>
<tr>
<th>IP Address</th>
<th>Ping Time (ms)</th>
<th>Last Successful</th>
</tr>
</thead>
<tbody>
{data.map((row, index) => (
<tr key={index}>
<td>{row.ip_address}</td>
<td>{row.ping_time}</td>
<td>{row.last_successful ? new Date(row.last_successful).toLocaleString() : "N/A"}</td>
</tr>
))}
</tbody>
</table>
);
}

export default Table;
